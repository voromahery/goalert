import React, { ReactElement, useState } from 'react'
import Button, { ButtonProps } from '@material-ui/core/Button'
import Grid from '@material-ui/core/Grid'
import Popover from '@material-ui/core/Popover'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core'
import { DateTime } from 'luxon'

const useStyles = makeStyles({
  buttonContainer: {
    display: 'flex',
    justifyContent: 'space-between',
    flex: 1,
  },
  paper: {
    padding: 8,
    maxWidth: 275,
  },
})

interface ActionBtnProps extends ButtonProps {
  onClick: NonNullable<ButtonProps['onClick']>
  children: NonNullable<ButtonProps['children']>
  title: NonNullable<ButtonProps['title']>
}

interface CalendarEventWrapperProps {
  children: ReactElement
  event: {
    start: Date
    end: Date
  }
  actions?: ActionBtnProps[]
}

export default function CalendarEventWrapper({
  children,
  event,
  actions,
}: CalendarEventWrapperProps): JSX.Element | null {
  const classes = useStyles()
  const [anchorEl, setAnchorEl] = useState<Element | null>(null)
  const open = Boolean(anchorEl)
  const id = open ? 'shift-popover' : undefined
  const fmt = (date: Date): string =>
    DateTime.fromJSDate(date).toLocaleString(DateTime.DATETIME_FULL)

  if (!children) return null
  return (
    <React.Fragment>
      <Popover
        id={id}
        open={open}
        anchorEl={anchorEl}
        onClose={() => setAnchorEl(null)}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'left',
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'left',
        }}
        PaperProps={
          {
            'data-cy': 'shift-tooltip',
          } as any
        }
        classes={{
          paper: classes.paper,
        }}
      >
        <Grid container spacing={1}>
          <Grid item xs={12}>
            <Typography variant='body2'>
              {`${fmt(event.start)}  –  ${fmt(event.end)}`}
            </Typography>
          </Grid>
          {actions?.length !== 0 && (
            <Grid item className={classes.buttonContainer}>
              {actions?.map((actionProps, i) => (
                <Button
                  key={i}
                  size='small'
                  variant='contained'
                  color='primary'
                  {...actionProps}
                  onClick={(e) => {
                    actionProps.onClick(e)
                    setAnchorEl(null)
                  }}
                />
              ))}
            </Grid>
          )}
        </Grid>
      </Popover>
      {React.cloneElement(children, {
        tabIndex: 0,
        onClick: (e: MouseEvent) => setAnchorEl(e.currentTarget as Element),
        onKeyDown: (e: KeyboardEvent) => {
          if (['Enter', ' '].includes(e.key)) {
            setAnchorEl(e.currentTarget as Element)
          }
        },
        role: 'button',
        'aria-pressed': open,
        'aria-describedby': id,
      })}
    </React.Fragment>
  )
}
