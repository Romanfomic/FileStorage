import { Component, inject } from '@angular/core';
import { UserService } from '../../services/user.service';
import { User } from '../../interfaces/user';
import { FormBuilder, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { TableModule } from 'primeng/table'
import { Dialog } from 'primeng/dialog';
import { Button } from 'primeng/button';
import { Observable, tap } from 'rxjs';
import { AsyncPipe, NgIf } from '@angular/common';
import { RoleService } from '../../services/role.service';
import { GroupService } from '../../services/group.service';
import { DropdownModule } from 'primeng/dropdown';

@Component({
    selector: 'app-users',
    imports: [
        TableModule,
        Dialog,
        FormsModule,
        ReactiveFormsModule,
        Button,
        NgIf,
        AsyncPipe,
        DropdownModule,
    ],
    templateUrl: './users.component.html',
    styleUrl: './users.component.less'
})
export class UsersComponent {
    private fb = inject(FormBuilder);
    private userService = inject(UserService);
    private roleService = inject(RoleService);
    private groupService = inject(GroupService);

    users: User[] = [];
    selectedUser: User | null = null;

    roles: { role_id: number, name: string }[] = [];
    groups: { group_id: number, name: string }[] = [];

    displayDialog = false;
    isNewUser = false;

    load$: Observable<any> = this.userService.getAllUsers().pipe(
        tap((users) => this.users = users)
    );
    loadRoles$: Observable<any> = this.roleService.getRoles().pipe(
        tap((roles) => this.roles = roles)
    );
    loadGroups$: Observable<any> = this.groupService.getAllGroups().pipe(
        tap((groups) => this.groups = groups)
    );
    
    userForm = this.fb.group({
        user_id: [null as number | null],
        login: ['', Validators.required],
        password: ['', this.isNewUser ? Validators.required : []],
        name: ['', Validators.required],
        surname: ['', Validators.required],
        mail: ['', [Validators.required, Validators.email]],
        role_id: [null as number | null],
        group_id: [null as number | null],
    });

    getRoleById(id: number): string {
        return this.roles.find((role) => role.role_id === id)?.name || '-';
    }

    getGroupById(id: number): string {
        return this.groups.find((group) => group.group_id === id)?.name || '-';
    }

    loadUsers(): void {
        this.load$ = this.userService.getAllUsers().pipe(
            tap((users) => this.users = users)
        )
    }

    openNew(): void {
        this.userForm.reset();
        this.isNewUser = true;
        this.displayDialog = true;
        
        this.userForm.get('password')?.setValidators(Validators.required);
        this.userForm.get('password')?.updateValueAndValidity();
    }

    editUser(user: User): void {
        this.userForm.patchValue(user);
        this.isNewUser = false;
        this.displayDialog = true;

        this.userForm.get('password')?.clearValidators();
        this.userForm.get('password')?.updateValueAndValidity();
    }

    deleteUser(user: User): void {
        this.load$ = this.userService.deleteUser(user.user_id!).pipe(
            tap(() => this.loadUsers())
        )
    }

    saveUser(): void {
        if (this.userForm.valid) {
            const userValue = this.userForm.value;
    
            if (this.isNewUser) {
                const newUser: Partial<User> & { password: string } = {
                    ...userValue,
                    password: userValue.password!,
                };
    
                this.load$ = this.userService.createUser(newUser).pipe(
                    tap(() => {
                        this.loadUsers();
                        this.displayDialog = false;
                    })
                )
            } else {
                const { password, ...updatedUser } = userValue;

                this.load$ = this.userService.updateUser(updatedUser as User).pipe(
                    tap(() => {
                        this.loadUsers();
                        this.displayDialog = false;
                    })
                )
            }
        }
    }
    
}
