import { Routes } from '@angular/router';
import { WrapComponent } from './components/wrap/wrap.component';
import { SharedStorageComponent } from './pages/shared-storage/shared-storage.component';

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
                path: '',
                pathMatch: 'full',
                redirectTo: 'storage'
            },
            {
                path: 'storage',
                loadComponent: () => import('./pages/storage/storage.component').then((m) => m.StorageComponent),
            },
            {
                path: 'shared',
                component: SharedStorageComponent
            },
            {
                path: 'users',
                loadComponent: () => import('./pages/users/users.component').then((m) => m.UsersComponent),
            },
            {
                path: 'groups',
                loadComponent: () => import('./pages/groups/groups.component').then((m) => m.GroupsComponent),
            },
            {
                path: 'roles',
                loadComponent: () => import('./pages/roles/roles.component').then((m) => m.RolesComponent),
            },
        ]
    },
    {
        path: '**',
        redirectTo: 'storage'
    }
];
