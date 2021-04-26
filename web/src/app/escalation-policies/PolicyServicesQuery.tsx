import React from 'react'
import { gql, useQuery } from '@apollo/client'
import PolicyServicesCard from './PolicyServicesCard'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'
import { Query } from '../../schema'

const query = gql`
  query($id: ID!) {
    escalationPolicy(id: $id) {
      id
      assignedTo {
        id
        name
      }
    }
  }
`

interface PolicyServicesQueryProps {
  escalationPolicyID: string
}

function PolicyServicesQuery(props: PolicyServicesQueryProps): JSX.Element {
  const { data, loading, error } = useQuery<Query>(query, {
    variables: { id: props.escalationPolicyID },
  })

  if (!data && loading) {
    return <Spinner />
  }

  if (error) {
    return <GenericError error={error.message} />
  }

  if (!data?.escalationPolicy) {
    return <ObjectNotFound />
  }

  return (
    <PolicyServicesCard services={data?.escalationPolicy.assignedTo || []} />
  )
}

export default PolicyServicesQuery
