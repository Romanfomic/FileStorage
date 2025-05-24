import { Component, inject, Optional } from '@angular/core';
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
import { TreeModule, TreeNodeDropEvent } from 'primeng/tree';
import { TreeDragDropService, TreeNode } from 'primeng/api';

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
        TreeModule,
    ],
    providers: [TreeDragDropService],
})
export class GroupsComponent {
    private fb = inject(FormBuilder);
    private groupService = inject(GroupService);
    @Optional() private dragDropService = inject(TreeDragDropService);

    groups: TreeNode[] = [];
    selectedParentid: number | null = null;
    displayDialog = false;
    isNewGroup = false;

    load$: Observable<any> = this.groupService.getGroups().pipe(
        tap((groups) => this.groups = groups.map(group => this.mapGroupToTreeNode(group)))
    );

    groupForm = this.fb.group({
        group_id: [null as number | null],
        name: ['', Validators.required],
        description: [''],
        parent_id: [null as number | null | undefined],
    });

    loadGroups(): void {
        this.load$ = this.groupService.getGroups().pipe(
            tap((groups) => this.groups = groups.map(group => this.mapGroupToTreeNode(group)))
        );
    }

    openNew(parentId?: number, event?: MouseEvent): void {
        event?.stopPropagation();
        this.selectedParentid = parentId ?? null;

        this.groupForm.reset();
        this.isNewGroup = true;
        this.displayDialog = true;
    }

    editGroup(group: Group, event?: MouseEvent): void {
        this.selectedParentid = group.parent_id ?? null;
        this.groupForm.patchValue(group);
        this.isNewGroup = false;
        this.displayDialog = true;
    }

    onDrop(event: any) {
        const childGroup = event.dragNode.data;
        const parentGroup = event.dropNode.data;

        childGroup.parent_id = parentGroup.group_id;

        if (event.dropNode.expanded) {
            childGroup.parent_id = parentGroup.parent_id;
        }

        this.load$ = this.groupService.updateGroup(childGroup.group_id!, childGroup).pipe(
            tap(() => {
                this.loadGroups();
                this.displayDialog = false;
            })
        );
        
        event.accept();
    }

    deleteGroup(group: Group, event?: MouseEvent): void {
        event?.stopPropagation();
        this.load$ = this.groupService.deleteGroup(group.group_id).pipe(
            tap(() => this.loadGroups())
        );
    }

    saveGroup(): void {
        if (this.groupForm.valid) {
            this.groupForm.patchValue({
                parent_id: this.selectedParentid
            });
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

    private mapGroupToTreeNode(group: Group): TreeNode {
        return {
            label: group.name,
            key: group.group_id.toString(),
            data: group,
            children: group.children?.map(child => this.mapGroupToTreeNode(child)) || []
        };
    }
}
