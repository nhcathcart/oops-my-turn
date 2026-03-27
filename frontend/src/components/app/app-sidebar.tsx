import { LogOut } from 'lucide-react'
import { LOGO_PATH, LOGO_VIEWBOX } from '@/lib/logo'
import { NavLink } from 'react-router'
import { useAuth } from '@/providers/auth-provider'
import { useLogout } from '@/api/auth'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'

const navItems = [{ title: 'Home', url: '/' }]

export function AppSidebar(props: React.ComponentProps<typeof Sidebar>) {
  const { user } = useAuth()
  const { logout } = useLogout()

  return (
    <Sidebar {...props}>
      <SidebarHeader className="px-4 py-4">
        <div className="flex items-center gap-2">
          <svg
            className="h-5 w-6"
            viewBox={LOGO_VIEWBOX}
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path d={LOGO_PATH} stroke="currentColor" strokeWidth="4" />
          </svg>
          <span className="text-lg font-semibold">oops-my-turn</span>
        </div>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {navItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <NavLink to={item.url} end>
                    {({ isActive }) => (
                      <SidebarMenuButton isActive={isActive}>{item.title}</SidebarMenuButton>
                    )}
                  </NavLink>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <div className="flex items-center gap-2 px-2 py-1.5 text-sm">
              <div className="flex min-w-0 flex-1 flex-col">
                <span className="truncate font-medium">
                  {user?.first_name} {user?.last_name}
                </span>
                <span className="text-muted-foreground truncate text-xs">{user?.email}</span>
              </div>
            </div>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <SidebarMenuButton onClick={logout}>
              <LogOut />
              Sign out
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  )
}
