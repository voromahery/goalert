package app

import (
	"context"
	"crypto/tls"
	"database/sql"
	"net"
	"net/http"

	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/app/lifecycle"
	"github.com/target/goalert/auth"
	"github.com/target/goalert/auth/basic"
	"github.com/target/goalert/auth/nonce"
	"github.com/target/goalert/calendarsubscription"
	"github.com/target/goalert/config"
	"github.com/target/goalert/engine"
	"github.com/target/goalert/engine/resolver"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/graphql"
	"github.com/target/goalert/graphql2/graphqlapp"
	"github.com/target/goalert/heartbeat"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/label"
	"github.com/target/goalert/limit"
	"github.com/target/goalert/notice"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/override"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/service"
	"github.com/target/goalert/timezone"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/user/favorite"
	"github.com/target/goalert/user/notificationrule"
	"github.com/target/goalert/util/sqlutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
)

// App represents an instance of the GoAlert application.
type App struct {
	cfg Config

	mgr *lifecycle.Manager

	db     *sql.DB
	l      net.Listener
	events *sqlutil.Listener

	cooldown *cooldown
	doneCh   chan struct{}

	sysAPIL   net.Listener
	sysAPISrv *grpc.Server
	hSrv      *health.Server

	srv         *http.Server
	requestLock *contextLocker
	startupErr  error

	notificationManager *notification.Manager
	Engine              *engine.Engine
	graphql             *graphql.Handler
	graphql2            *graphqlapp.App
	AuthHandler         *auth.Handler

	twilioSMS    *twilio.SMS
	twilioVoice  *twilio.Voice
	twilioConfig *twilio.Config

	slackChan *slack.ChannelSender

	ConfigStore *config.Store

	AlertStore    alert.Store
	AlertLogStore alertlog.Store

	AuthBasicStore        *basic.Store
	UserStore             user.Store
	ContactMethodStore    contactmethod.Store
	NotificationRuleStore notificationrule.Store
	FavoriteStore         favorite.Store

	ServiceStore        service.Store
	EscalationStore     escalation.Store
	IntegrationKeyStore integrationkey.Store
	ScheduleRuleStore   rule.Store
	NotificationStore   notification.Store
	ScheduleStore       *schedule.Store
	RotationStore       rotation.Store

	CalSubStore    *calendarsubscription.Store
	OverrideStore  override.Store
	Resolver       resolver.Resolver
	LimitStore     *limit.Store
	HeartbeatStore heartbeat.Store

	OAuthKeyring   keyring.Keyring
	SessionKeyring keyring.Keyring
	APIKeyring     keyring.Keyring

	NonceStore    nonce.Store
	LabelStore    label.Store
	OnCallStore   oncall.Store
	NCStore       notificationchannel.Store
	TimeZoneStore *timezone.Store
	NoticeStore   *notice.Store
}

// NewApp constructs a new App and binds the listening socket.
func NewApp(c Config, db *sql.DB) (*App, error) {
	l, err := net.Listen("tcp", c.ListenAddr)
	if err != nil {
		return nil, errors.Wrapf(err, "bind address %s", c.ListenAddr)
	}

	if c.TLSListenAddr != "" {
		l2, err := tls.Listen("tcp", c.TLSListenAddr, c.TLSConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "listen %s", c.TLSListenAddr)
		}
		l = newMultiListener(l, l2)
	}

	app := &App{
		l:      l,
		db:     db,
		cfg:    c,
		doneCh: make(chan struct{}),

		requestLock: newContextLocker(),
	}
	if c.KubernetesCooldown > 0 {
		app.cooldown = newCooldown(c.KubernetesCooldown)
	}

	if c.StatusAddr != "" {
		err = listenStatus(c.StatusAddr, app.doneCh)
		if err != nil {
			return nil, errors.Wrap(err, "start status listener")
		}
	}

	app.db.SetMaxIdleConns(c.DBMaxIdle)
	app.db.SetMaxOpenConns(c.DBMaxOpen)

	app.mgr = lifecycle.NewManager(app._Run, app._Shutdown)
	err = app.mgr.SetStartupFunc(app.startup)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// WaitForStartup will wait until the startup sequence is completed or the context is expired.
func (a *App) WaitForStartup(ctx context.Context) error { return a.mgr.WaitForStartup(ctx) }

// DB returns the sql.DB instance used by the application.
func (a *App) DB() *sql.DB { return a.db }

// URL returns the non-TLS listener URL of the application.
func (a *App) URL() string {
	return "http://" + a.l.Addr().String()
}

// Status returns the current lifecycle status of the App.
func (a *App) Status() lifecycle.Status {
	return a.mgr.Status()
}

// ActiveRequests returns the current number of active
// requests, not including pending ones during pause.
func (a *App) ActiveRequests() int {
	return a.requestLock.RLockCount()
}
