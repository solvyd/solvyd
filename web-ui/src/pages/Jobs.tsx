import { useQuery } from '@tanstack/react-query'
import { jobsApi } from '@/lib/api'
import { Job } from '@/types'
import { Link } from 'react-router-dom'
import { Plus, Play, Settings } from 'lucide-react'

export default function Jobs() {
  const { data, isLoading } = useQuery({
    queryKey: ['jobs'],
    queryFn: () => jobsApi.list(),
  })

  const jobs: Job[] = data?.data || []

  if (isLoading) {
    return <div>Loading...</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Jobs</h1>
        <button className="flex items-center px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90">
          <Plus className="h-5 w-5 mr-2" />
          New Job
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {jobs.map((job) => (
          <div key={job.id} className="bg-white rounded-lg shadow p-6">
            <div className="flex items-start justify-between">
              <div>
                <Link 
                  to={`/jobs/${job.id}`}
                  className="text-lg font-semibold text-gray-900 hover:text-primary"
                >
                  {job.name}
                </Link>
                <p className="mt-1 text-sm text-gray-500">{job.description}</p>
              </div>
              <div className={`px-2 py-1 text-xs font-medium rounded ${
                job.enabled 
                  ? 'bg-green-100 text-green-800' 
                  : 'bg-gray-100 text-gray-800'
              }`}>
                {job.enabled ? 'Enabled' : 'Disabled'}
              </div>
            </div>

            <div className="mt-4 space-y-2">
              <div className="flex items-center text-sm text-gray-600">
                <span className="font-medium mr-2">SCM:</span>
                <span>{job.scm_type}</span>
              </div>
              <div className="flex items-center text-sm text-gray-600">
                <span className="font-medium mr-2">Branch:</span>
                <span>{job.scm_branch}</span>
              </div>
            </div>

            <div className="mt-6 flex space-x-2">
              <button 
                onClick={() => jobsApi.trigger(job.id)}
                className="flex-1 flex items-center justify-center px-3 py-2 bg-primary text-white text-sm font-medium rounded hover:bg-primary/90"
              >
                <Play className="h-4 w-4 mr-2" />
                Trigger
              </button>
              <Link
                to={`/jobs/${job.id}`}
                className="flex items-center justify-center px-3 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                <Settings className="h-4 w-4" />
              </Link>
            </div>
          </div>
        ))}
      </div>

      {jobs.length === 0 && (
        <div className="text-center py-12">
          <p className="text-gray-500">No jobs configured yet</p>
          <button className="mt-4 flex items-center mx-auto px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90">
            <Plus className="h-5 w-5 mr-2" />
            Create Your First Job
          </button>
        </div>
      )}
    </div>
  )
}
