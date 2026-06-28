import { useTranslation } from 'react-i18next'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { useAuthStore } from '@/store/auth'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

interface SidebarProps {
  user: {
    email: string
    username: string
    role: number
  }
}

export function Sidebar({ user }: SidebarProps) {
  const { t } = useTranslation()
  const location = useLocation()
  const navigate = useNavigate()
  const clearTokens = useAuthStore((state) => state.clearTokens)

  const handleLogout = () => {
    clearTokens()
    navigate('/login')
  }

  const userLinks = [
    { path: '/dashboard', label: t('sidebar.dashboard'), icon: '📊' },
    { path: '/keys', label: t('sidebar.apiKeys'), icon: '🔑' },
    { path: '/models', label: t('sidebar.models'), icon: '🧠' },
    { path: '/usage', label: t('sidebar.usage'), icon: '📈' },
    { path: '/redeem', label: t('sidebar.redeem'), icon: '🎫' },
  ]

  const adminLinks = [
    { path: '/admin/dashboard', label: t('sidebar.adminDashboard'), icon: '📊' },
    { path: '/admin/users', label: t('sidebar.users'), icon: '👥' },
    { path: '/admin/channels', label: t('sidebar.channels'), icon: '🔗' },
    { path: '/admin/pricing', label: t('sidebar.pricing'), icon: '💰' },
    { path: '/admin/redeem-codes', label: t('sidebar.redeemCodes'), icon: '🎟️' },
    { path: '/admin/logs', label: t('sidebar.logs'), icon: '📝' },
  ]

  return (
    <aside className="h-screen max-h-screen w-64 shrink-0 bg-card border-r flex flex-col overflow-hidden">
      <div className="p-4 border-b">
        <div className="flex items-center gap-2.5">
          <img src="/favicon.svg" alt="AiToken" className="w-8 h-8 shrink-0" />
          <h2 className="text-lg font-bold leading-tight">AiToken分发站</h2>
        </div>
      </div>

      <nav className="flex-1 p-4 space-y-1 overflow-y-auto">
        {userLinks.map((link) => (
          <Link
            key={link.path}
            to={link.path}
            className={cn(
              "flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors",
              location.pathname === link.path
                ? "bg-primary text-primary-foreground"
                : "hover:bg-muted"
            )}
          >
            <span>{link.icon}</span>
            <span>{link.label}</span>
          </Link>
        ))}

        {user.role >= 10 && (
          <>
            <div className="pt-4 pb-2 px-3">
              <span className="text-xs font-semibold text-muted-foreground uppercase">
                {t('sidebar.admin')}
              </span>
            </div>
            {adminLinks.map((link) => (
              <Link
                key={link.path}
                to={link.path}
                className={cn(
                  "flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors",
                  location.pathname === link.path
                    ? "bg-primary text-primary-foreground"
                    : "hover:bg-muted"
                )}
              >
                <span>{link.icon}</span>
                <span>{link.label}</span>
              </Link>
            ))}
          </>
        )}
      </nav>

      <div className="p-4">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button className="flex items-center gap-3 w-full p-2 rounded-md hover:bg-muted transition-colors cursor-pointer outline-none">
              <div className="w-8 h-8 rounded-full bg-primary flex items-center justify-center text-primary-foreground text-sm font-medium">
                {user.username.charAt(0).toUpperCase()}
              </div>
              <div className="flex-1 min-w-0 text-left">
                <div className="text-sm font-medium truncate">{user.username}</div>
                <div className="text-xs text-muted-foreground truncate">{user.email}</div>
              </div>
              {user.role >= 10 && (
                <Badge variant="default" className="text-xs">{t('common.admin')}</Badge>
              )}
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-56" align="end" side="top">
            <DropdownMenuLabel>
              <div className="flex flex-col space-y-1">
                <p className="text-sm font-medium">{user.username}</p>
                <p className="text-xs text-muted-foreground">{user.email}</p>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => navigate('/settings')}>
              {t('common.settings')}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleLogout} className="text-destructive">
              {t('common.logout')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </aside>
  )
}
