import { Component, inject } from '@angular/core';
import { Router, RouterOutlet } from '@angular/router';
import { ButtonModule } from 'primeng/button';
import { AuthService } from '../../services/auth.service';

@Component({
    selector: 'app-wrap',
    imports: [RouterOutlet, ButtonModule],
    templateUrl: './wrap.component.html',
    styleUrl: './wrap.component.less'
})
export class WrapComponent {
    private router = inject(Router);
    private authService = inject(AuthService)

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
