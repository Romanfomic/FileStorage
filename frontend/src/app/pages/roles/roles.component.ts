import { Component, inject } from '@angular/core';
import { RoleService } from '../../services/role.service';
import { Role } from '../../interfaces/role';
import { FormBuilder, Validators, FormsModule, ReactiveFormsModule } from '@angular/forms';
import { TableModule } from 'primeng/table';
import { Dialog } from 'primeng/dialog';
import { Button } from 'primeng/button';
import { Observable, tap } from 'rxjs';
import { AsyncPipe, NgFor, NgIf } from '@angular/common';
import { InputTextModule } from 'primeng/inputtext';
import { PermissionService } from '../../services/permission.service';
import { Permission } from '../../interfaces/permission';
import { MultiSelectModule } from 'primeng/multiselect';
import { Textarea } from 'primeng/textarea';

@Component({
    selector: 'app-roles',
    standalone: true,
    imports: [
        TableModule,
        Dialog,
        Button,
        FormsModule,
        ReactiveFormsModule,
        NgIf,
        NgFor,
        AsyncPipe,
        InputTextModule,
        MultiSelectModule,
        Textarea,
    ],
    templateUrl: './roles.component.html',
    styleUrl: './roles.component.less'
})
export class RolesComponent {
    private fb = inject(FormBuilder);
    private roleService = inject(RoleService);
    private permissionService = inject(PermissionService);

    roles: Role[] = [];
    selectedRole: Role | null = null;
    permissions: Permission[] = [];

    displayDialog = false;
    isNewRole = false;

    roleForm = this.fb.group({
        role_id: [null as number | null],
        name: ['', Validators.required],
        description: [''],
        permissions: [[] as number[]]
    });

    private translatePermissions: { [key: string]: string } = {
        manage_users: 'Управление пользователями',
        manage_roles: 'Управление ролями',
        manage_groups: 'Управление группами',
    };

    load$: Observable<any> = this.roleService.getRoles().pipe(
        tap((roles) => this.roles = roles)
    );

    loadPermissions$: Observable<any> = this.permissionService.getPermissions().pipe(
        tap(perms => this.permissions = perms)
    );

    getPermissionName(id: number): string {
        const name = this.permissions.find(p => p.permission_id === id)?.name;

        return name ? this.translatePermissions[name] : '-';
    }

    openNew(): void {
        this.roleForm.reset();
        this.isNewRole = true;
        this.displayDialog = true;
    }

    editRole(role: Role): void {
        this.roleForm.patchValue({
            ...role,
            permissions: role.permissions ?? []
        });
        this.isNewRole = false;
        this.displayDialog = true;
    }

    deleteRole(role: Role): void {
        if (role.role_id != null) {
            this.load$ = this.roleService.deleteRole(role.role_id).pipe(
                tap(() => this.loadRoles())
            );
        }
    }

    saveRole(): void {
        if (this.roleForm.valid) {
            const role = this.roleForm.value as Role;

            if (this.isNewRole) {
                this.load$ = this.roleService.createRole(role).pipe(
                    tap(() => {
                        this.loadRoles();
                        this.displayDialog = false;
                    })
                );
            } else {
                this.load$ = this.roleService.updateRole(role.role_id!, role).pipe(
                    tap(() => {
                        this.loadRoles();
                        this.displayDialog = false;
                    })
                );
            }
        }
    }

    loadRoles(): void {
        this.load$ = this.roleService.getRoles().pipe(
            tap((roles) => this.roles = roles)
        );
    }
}
