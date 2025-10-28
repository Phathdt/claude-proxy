import { Link, Outlet, useNavigate, useLocation } from 'react-router-dom'
import { LayoutDashboard, Key, LogOut, Shield, User, BarChart3, Monitor } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuth } from '@/hooks/useAuth'
import { ThemeToggle } from '@/components/theme-toggle'

const navigation = [
  { name: 'Dashboard', href: '/admin/dashboard', icon: LayoutDashboard },
  { name: 'Statistics', href: '/admin/statistics', icon: BarChart3 },
  { name: 'Accounts', href: '/admin/accounts', icon: Shield },
  { name: 'Tokens', href: '/admin/tokens', icon: Key },
  { name: 'Sessions', href: '/admin/sessions', icon: Monitor },
]

export function AdminLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuth()

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  return (
    <div className="bg-background min-h-screen">
      {/* Sidebar */}
      <div className="bg-sidebar border-sidebar-border fixed inset-y-0 left-0 z-50 w-64 border-r">
        <div className="flex h-full flex-col">
          {/* Logo */}
          <div className="border-sidebar-border flex h-16 items-center border-b px-6">
            <h1 className="text-sidebar-foreground text-xl font-bold">Claude Proxy</h1>
          </div>

          {/* Navigation */}
          <nav className="flex-1 space-y-1 px-3 py-4">
            {navigation.map((item) => {
              const isActive = location.pathname === item.href
              return (
                <Link
                  key={item.name}
                  to={item.href}
                  className={cn(
                    'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                    isActive
                      ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                      : 'text-sidebar-foreground hover:bg-sidebar-accent/50'
                  )}
                >
                  <item.icon className="h-5 w-5" />
                  {item.name}
                </Link>
              )
            })}
          </nav>

          {/* User section */}
          <div className="border-sidebar-border border-t p-4">
            {user && (
              <div className="bg-sidebar-accent/30 mb-3 rounded-lg p-3">
                <div className="flex items-center gap-2">
                  <User className="text-sidebar-foreground h-4 w-4" />
                  <div className="min-w-0 flex-1">
                    <p className="text-sidebar-foreground truncate text-sm font-medium">
                      {user.name}
                    </p>
                    <p className="text-sidebar-foreground/70 truncate text-xs">{user.email}</p>
                  </div>
                </div>
              </div>
            )}
            <div className="mb-3 flex gap-2">
              <ThemeToggle />
            </div>
            <button
              onClick={handleLogout}
              className="text-sidebar-foreground hover:bg-sidebar-accent/50 flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors"
            >
              <LogOut className="h-5 w-5" />
              Logout
            </button>
          </div>
        </div>
      </div>

      {/* Main content */}
      <div className="pl-64">
        <main className="p-8">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
