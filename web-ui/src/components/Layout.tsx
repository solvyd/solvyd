import { Outlet, Link, useLocation } from 'react-router-dom'
import { 
  LayoutDashboard, 
  Briefcase, 
  Play, 
  Server, 
  Rocket, 
  Puzzle,
  Activity
} from 'lucide-react'

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Jobs', href: '/jobs', icon: Briefcase },
  { name: 'Builds', href: '/builds', icon: Play },
  { name: 'Workers', href: '/workers', icon: Server },
  { name: 'Deployments', href: '/deployments', icon: Rocket },
  { name: 'Plugins', href: '/plugins', icon: Puzzle },
]

export default function Layout() {
  const location = useLocation()

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Sidebar */}
      <div className="fixed inset-y-0 left-0 w-64 bg-white border-r border-gray-200">
        <div className="flex h-16 items-center px-6 border-b border-gray-200">
          <Activity className="h-8 w-8 text-primary" />
          <span className="ml-2 text-xl font-bold text-gray-900">Ritmo</span>
        </div>
        
        <nav className="mt-6 px-3">
          {navigation.map((item) => {
            const isActive = location.pathname === item.href
            return (
              <Link
                key={item.name}
                to={item.href}
                className={`
                  flex items-center px-3 py-2 mb-1 rounded-lg text-sm font-medium
                  transition-colors
                  ${isActive 
                    ? 'bg-primary text-white' 
                    : 'text-gray-700 hover:bg-gray-100'
                  }
                `}
              >
                <item.icon className="h-5 w-5 mr-3" />
                {item.name}
              </Link>
            )
          })}
        </nav>
      </div>

      {/* Main content */}
      <div className="pl-64">
        <div className="flex h-16 items-center px-6 border-b border-gray-200 bg-white">
          <h1 className="text-lg font-semibold text-gray-900">
            {navigation.find(nav => nav.href === location.pathname)?.name || 'Ritmo CI/CD'}
          </h1>
        </div>
        
        <main className="p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
