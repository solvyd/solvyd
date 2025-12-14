import { useQuery } from '@tanstack/react-query'
import { buildsApi, workersApi } from '@/lib/api'
import { Build, Worker } from '@/types'
import { Link } from 'react-router-dom'
import { 
  Activity, 
  CheckCircle2, 
  XCircle, 
  Clock, 
  Server,
  TrendingUp,
  TrendingDown
} from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'

export default function Dashboard() {
  const { data: buildsData } = useQuery({
    queryKey: ['builds', { limit: 10 }],
    queryFn: () => buildsApi.list({ limit: 10 }),
  })

  const { data: workersData } = useQuery({
    queryKey: ['workers'],
    queryFn: () => workersApi.list(),
  })

  const builds: Build[] = buildsData?.data || []
  const workers: Worker[] = workersData?.data || []

  // Calculate statistics
  const buildStats = {
    total: builds.length,
    running: builds.filter(b => b.status === 'running').length,
    queued: builds.filter(b => b.status === 'queued').length,
    success: builds.filter(b => b.status === 'success').length,
    failed: builds.filter(b => b.status === 'failed').length,
  }

  const workerStats = {
    total: workers.length,
    online: workers.filter(w => w.status === 'online').length,
    offline: workers.filter(w => w.status === 'offline').length,
    utilization: workers.reduce((acc, w) => acc + (w.current_builds / w.max_concurrent_builds), 0) / Math.max(workers.length, 1),
  }

  const successRate = buildStats.total > 0 
    ? ((buildStats.success / (buildStats.success + buildStats.failed)) * 100).toFixed(1)
    : '0'

  return (
    <div className="space-y-6">
      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Active Builds"
          value={buildStats.running}
          icon={<Activity className="h-6 w-6" />}
          trend="+12%"
          trendUp={true}
        />
        <StatCard
          title="Queued Builds"
          value={buildStats.queued}
          icon={<Clock className="h-6 w-6" />}
          trend="-5%"
          trendUp={false}
        />
        <StatCard
          title="Online Workers"
          value={`${workerStats.online}/${workerStats.total}`}
          icon={<Server className="h-6 w-6" />}
          trend="100%"
          trendUp={true}
        />
        <StatCard
          title="Success Rate"
          value={`${successRate}%`}
          icon={<CheckCircle2 className="h-6 w-6" />}
          trend="+3%"
          trendUp={true}
        />
      </div>

      {/* Recent Builds */}
      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Recent Builds</h2>
        </div>
        <div className="divide-y divide-gray-200">
          {builds.length === 0 ? (
            <div className="px-6 py-12 text-center text-gray-500">
              No builds yet
            </div>
          ) : (
            builds.slice(0, 10).map((build) => (
              <Link
                key={build.id}
                to={`/builds/${build.id}`}
                className="block px-6 py-4 hover:bg-gray-50 transition-colors"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    <BuildStatusIcon status={build.status} />
                    <div>
                      <div className="font-medium text-gray-900">
                        {build.job_name || build.job_id} #{build.build_number}
                      </div>
                      <div className="text-sm text-gray-500">
                        {build.scm_commit_message?.substring(0, 60)}...
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm font-medium text-gray-900">
                      {build.status}
                    </div>
                    <div className="text-sm text-gray-500">
                      {formatDistanceToNow(new Date(build.queued_at), { addSuffix: true })}
                    </div>
                  </div>
                </div>
              </Link>
            ))
          )}
        </div>
      </div>

      {/* Workers Status */}
      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Workers</h2>
        </div>
        <div className="divide-y divide-gray-200">
          {workers.length === 0 ? (
            <div className="px-6 py-12 text-center text-gray-500">
              No workers registered
            </div>
          ) : (
            workers.map((worker) => (
              <div key={worker.id} className="px-6 py-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    <div className={`w-3 h-3 rounded-full ${
                      worker.status === 'online' ? 'bg-green-500' : 'bg-gray-300'
                    }`} />
                    <div>
                      <div className="font-medium text-gray-900">{worker.name}</div>
                      <div className="text-sm text-gray-500">
                        {worker.current_builds}/{worker.max_concurrent_builds} builds
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm text-gray-500">
                      {worker.cpu_cores} cores, {(worker.memory_mb / 1024).toFixed(0)}GB RAM
                    </div>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}

function StatCard({ 
  title, 
  value, 
  icon, 
  trend, 
  trendUp 
}: { 
  title: string
  value: string | number
  icon: React.ReactNode
  trend: string
  trendUp: boolean
}) {
  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-600">{title}</p>
          <p className="mt-2 text-3xl font-bold text-gray-900">{value}</p>
        </div>
        <div className="p-3 bg-primary/10 rounded-lg text-primary">
          {icon}
        </div>
      </div>
      <div className="mt-4 flex items-center">
        {trendUp ? (
          <TrendingUp className="h-4 w-4 text-green-500" />
        ) : (
          <TrendingDown className="h-4 w-4 text-red-500" />
        )}
        <span className={`ml-2 text-sm ${trendUp ? 'text-green-600' : 'text-red-600'}`}>
          {trend}
        </span>
        <span className="ml-2 text-sm text-gray-500">from last week</span>
      </div>
    </div>
  )
}

function BuildStatusIcon({ status }: { status: string }) {
  switch (status) {
    case 'success':
      return <CheckCircle2 className="h-5 w-5 text-green-500" />
    case 'failed':
      return <XCircle className="h-5 w-5 text-red-500" />
    case 'running':
      return <Activity className="h-5 w-5 text-blue-500 animate-pulse" />
    case 'queued':
      return <Clock className="h-5 w-5 text-yellow-500" />
    default:
      return <div className="h-5 w-5 rounded-full bg-gray-300" />
  }
}
