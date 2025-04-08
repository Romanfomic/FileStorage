import { Routes } from '@angular/router';
import { WrapComponent } from './components/wrap/wrap.component';

export const routes: Routes = [
    {
        path: 'auth',
        loadComponent: () => import('./pages/auth/auth.component').then((c) => c.AuthComponent),
    },
    {
        path: '',
        component: WrapComponent,
        children: [
            {
                path: 'storage',
                loadComponent: () => import('./pages/storage/storage.component').then((m) => m.StorageComponent),
            },
            {
                path: 'users',
                loadComponent: () => import('./pages/users/users.component').then((m) => m.UsersComponent),
            },
        ]
    },
    {
        path: '**',
        redirectTo: 'storage'
    }
];
