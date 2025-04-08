import { Component, inject } from '@angular/core';
import { FormBuilder, Validators } from '@angular/forms';
import { GroupService } from '../../services/group.service';
import { Group } from '../../interfaces/group';
import { Observable, tap } from 'rxjs';
import { TableModule } from 'primeng/table';
import { Dialog } from 'primeng/dialog';
import { Button } from 'primeng/button';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { AsyncPipe, NgIf } from '@angular/common';
import { InputTextModule } from 'primeng/inputtext';
import { TextareaModule } from 'primeng/textarea';

@Component({
    selector: 'app-groups',
    standalone: true,
    templateUrl: './groups.component.html',
    styleUrl: './groups.component.less',
    imports: [
        TableModule,
        Dialog,
        FormsModule,
        ReactiveFormsModule,
        Button,
        NgIf,
        AsyncPipe,
        InputTextModule,
        TextareaModule,
    ],
})
export class GroupsComponent {
    private fb = inject(FormBuilder);
    private groupService = inject(GroupService);

    groups: Group[] = [];
    selectedGroup: Group | null = null;
    displayDialog = false;
    isNewGroup = false;

    load$: Observable<any> = this.groupService.getGroups().pipe(
        tap((groups) => this.groups = groups)
    );

    groupForm = this.fb.group({
        group_id: [null as number | null],
        name: ['', Validators.required],
        description: [''],
    });

    loadGroups(): void {
        this.load$ = this.groupService.getGroups().pipe(
            tap((groups) => this.groups = groups)
        );
    }

    openNew(): void {
        this.groupForm.reset();
        this.isNewGroup = true;
        this.displayDialog = true;
    }

    editGroup(group: Group): void {
        this.groupForm.patchValue(group);
        this.isNewGroup = false;
        this.displayDialog = true;
    }

    deleteGroup(group: Group): void {
        this.load$ = this.groupService.deleteGroup(group.group_id).pipe(
            tap(() => this.loadGroups())
        );
    }

    saveGroup(): void {
        if (this.groupForm.valid) {
            const groupValue = this.groupForm.value as Group;

            if (this.isNewGroup) {
                this.load$ = this.groupService.createGroup(groupValue).pipe(
                    tap(() => {
                        this.loadGroups();
                        this.displayDialog = false;
                    })
                );
            } else {
                this.load$ = this.groupService.updateGroup(groupValue.group_id!, groupValue).pipe(
                    tap(() => {
                        this.loadGroups();
                        this.displayDialog = false;
                    })
                );
            }
        }
    }
}
