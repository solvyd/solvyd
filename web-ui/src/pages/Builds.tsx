import { useQuery } from '@tanstack/react-query'
import { buildsApi } from '@/lib/api'
import { Build } from '@/types'
import { Link } from 'react-router-dom'
import { formatDistanceToNow } from 'date-fns'
import { CheckCircle2, XCircle, Activity, Clock, X } from 'lucide-react'

export default function Builds() {
  const { data, isLoading } = useQuery({
    queryKey: ['builds'],
    queryFn: () => buildsApi.list({ limit: 50 }),
  })

  const builds: Build[] = data?.data || []

  if (isLoading) {
    return <div>Loading...</div>
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">Builds</h1>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Build
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Commit
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Duration
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Triggered
              </th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {builds.map((build) => (
              <tr key={build.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap">
                  <Link 
                    to={`/builds/${build.id}`}
                    className="text-sm font-medium text-primary hover:text-primary/80"
                  >
                    {build.job_name || 'Unknown'} #{build.build_number}
                  </Link>
                  <div className="text-sm text-gray-500">{build.branch}</div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <BuildStatusBadge status={build.status} />
                </td>
                <td className="px-6 py-4">
                  <div className="text-sm text-gray-900">
                    {build.scm_commit_sha?.substring(0, 7)}
                  </div>
                  <div className="text-sm text-gray-500 truncate max-w-xs">
                    {build.scm_commit_message}
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {build.duration_seconds 
                    ? `${Math.floor(build.duration_seconds / 60)}m ${build.duration_seconds % 60}s`
                    : '-'
                  }
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {formatDistanceToNow(new Date(build.queued_at), { addSuffix: true })}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  {(build.status === 'running' || build.status === 'queued') && (
                    <button
                      onClick={() => buildsApi.cancel(build.id)}
                      className="text-red-600 hover:text-red-900"
                    >
                      Cancel
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {builds.length === 0 && (
          <div className="text-center py-12">
            <p className="text-gray-500">No builds found</p>
          </div>
        )}
      </div>
    </div>
  )
}

function BuildStatusBadge({ status }: { status: string }) {
  const config = {
    success: { icon: CheckCircle2, color: 'green', label: 'Success' },
    failed: { icon: XCircle, color: 'red', label: 'Failed' },
    running: { icon: Activity, color: 'blue', label: 'Running' },
    queued: { icon: Clock, color: 'yellow', label: 'Queued' },
    cancelled: { icon: X, color: 'gray', label: 'Cancelled' },
  }[status] || { icon: Clock, color: 'gray', label: status }

  const Icon = config.icon

  return (
    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-${config.color}-100 text-${config.color}-800`}>
      <Icon className="h-4 w-4 mr-1" />
      {config.label}
    </span>
  )
}
