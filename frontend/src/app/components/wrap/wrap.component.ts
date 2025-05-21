import { Component, inject } from '@angular/core';
import { Router, RouterOutlet } from '@angular/router';
import { ButtonModule } from 'primeng/button';
import { AuthService } from '../../services/auth.service';
import { TokenPayload } from '../../interfaces/token';
import { jwtDecode } from "jwt-decode";

@Component({
    selector: 'app-wrap',
    imports: [RouterOutlet, ButtonModule],
    templateUrl: './wrap.component.html',
    styleUrl: './wrap.component.less'
})
export class WrapComponent {
    private router = inject(Router);
    private authService = inject(AuthService)

    get isUsersManage(): boolean {
      return this.checkPermissions("manage_users");
    }

    get isGroupsManage(): boolean {
      return this.checkPermissions("manage_groups");
    }

    get isRolesManage(): boolean {
      return this.checkPermissions("manage_roles");
    }

    checkPermissions(permission: string): boolean {
      const token = localStorage.getItem('token');
      if (!token) return false;

      const decoded = jwtDecode<TokenPayload>(token);
      if (decoded.type === 'admin') return true;
      if (decoded.permissions && decoded.permissions.length && decoded.permissions.includes(permission)) return true
      return false;
    }

    navigate(path: string) {
        this.router.navigate([path])
    }

    logout() {
      this.authService.logout();
      this.router.navigate(['/auth']);
    }
  
    getTitle(): string {
      const route = this.router.url.split('/')[1];
      const titles: Record<string, string> = {
            storage: 'Хранилище',
            shared: 'Доступные мне',
            users: 'Пользователи',
            roles: 'Роли',
            groups: 'Группы',
      };
      
      return titles[route] || '';
    }
}
