import React from 'react'
import Card from '@material-ui/core/Card'
import FlatList, { FlatListListItem } from '../lists/FlatList'
import { Target } from '../../schema'
import _ from 'lodash'

interface PolicyServicesCardProps {
  services: Target[]
}

function PolicyServicesCard(props: PolicyServicesCardProps): JSX.Element {
  function getServicesItems(): FlatListListItem[] {
    return _.sortBy(
      props.services.map((service) => ({
        title: service.name ?? '',
        url: `/services/${service.id}`,
      })),
      [(svc) => svc.title],
    )
  }

  return (
    <Card>
      <FlatList
        emptyMessage='No services are associated with this Escalation Policy.'
        items={getServicesItems()}
      />
    </Card>
  )
}

export default PolicyServicesCard
