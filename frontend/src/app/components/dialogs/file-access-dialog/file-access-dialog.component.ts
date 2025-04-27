import { Component, EventEmitter, inject, Input, Output } from '@angular/core';
import { FileMetadata } from '../../../interfaces/fileData';
import { User } from '../../../interfaces/user';
import { FileService } from '../../../services/file.service';
import { UserService } from '../../../services/user.service';
import { TabsModule } from 'primeng/tabs';
import { CheckboxModule } from 'primeng/checkbox';
import { DropdownModule } from 'primeng/dropdown';
import { ButtonModule } from 'primeng/button';
import { DialogModule } from 'primeng/dialog';
import { FormsModule } from '@angular/forms';
import { AsyncPipe, NgFor, NgIf } from '@angular/common';
import { Observable, tap } from 'rxjs';
import { GroupService } from '../../../services/group.service';
import { Group } from '../../../interfaces/group';

@Component({
    selector: 'app-file-access-dialog',
    imports: [
        TabsModule, 
        CheckboxModule,
        DropdownModule,
        ButtonModule,
        DialogModule,
        FormsModule,
        NgFor,
        NgIf,
        AsyncPipe,
    ],
    templateUrl: './file-access-dialog.component.html',
    styleUrl: './file-access-dialog.component.less'
})
export class FileAccessDialogComponent {
    private fileService = inject(FileService);
    private userService = inject(UserService);
    private groupService = inject(GroupService);

    @Input() visible = false;
    @Output() visibleChange = new EventEmitter<boolean>();
    
    file!: FileMetadata;

    allUsers: User[] = [];
    allGroups: Group[] = [];

    userAccessLevelMap: { [userId: number]: number | null } = {};
    groupAccessLevelMap: { [groupId: number]: number | null } = {};
    initialUserAccessLevelMap: { [userId: number]: number | null } = {};
    initialGroupAccessLevelMap: { [groupId: number]: number | null } = {};

    fullAccessOptions = [
        { label: 'Нет доступа', value: null },
        { label: 'Чтение', value: 1 },
        { label: 'Чтение и запись', value: 2 },
    ];
    
    load$!: Observable<any>;
    loadUsers$!: Observable<any>;
    loadGroups$!: Observable<any>;

    open(file: FileMetadata) {
        this.file = file;
        this.visible = true;
    
        this.loadUsers$ = this.userService.getAllUsers().pipe(
            tap((users) => this.allUsers = users)
        );
    
        this.loadGroups$ = this.groupService.getGroups().pipe(
            tap((groups) => this.allGroups = groups)
        );
    
        this.load$ = this.fileService.getFilePermissions(this.file.file_id).pipe(
            tap((res) => {
                this.userAccessLevelMap = {};
                this.groupAccessLevelMap = {};
                this.initialUserAccessLevelMap = {};
                this.initialGroupAccessLevelMap = {};
    
                if (!!res.users) {
                    for (let u of res.users) {
                        this.userAccessLevelMap[u.user_id] = u.access_id;
                        this.initialUserAccessLevelMap[u.user_id] = u.access_id;
                    }
                }
    
                if (!!res.groups) {
                    for (let g of res.groups) {
                        this.groupAccessLevelMap[g.group_id] = g.access_id;
                        this.initialGroupAccessLevelMap[g.group_id] = g.access_id;
                    }
                }
            }),
        );
    }

    save() {
        const fileId = this.file.file_id;
    
        // Сохраняем изменения для пользователей
        for (let userIdStr in this.userAccessLevelMap) {
            const userId = Number(userIdStr);
            const newAccess = this.userAccessLevelMap[userId];
            const initialAccess = this.initialUserAccessLevelMap[userId];
    
            if (newAccess !== initialAccess) { // Только если изменилось
                if (newAccess === null || newAccess === undefined) {
                    this.fileService.revokeUserAccess(fileId, userId).subscribe();
                } else {
                    this.fileService.shareFileWithUser(fileId, userId, newAccess).subscribe();
                }
            }
        }
    
        // Сохраняем изменения для групп
        for (let groupIdStr in this.groupAccessLevelMap) {
            const groupId = Number(groupIdStr);
            const newAccess = this.groupAccessLevelMap[groupId];
            const initialAccess = this.initialGroupAccessLevelMap[groupId];
    
            if (newAccess !== initialAccess) { // Только если изменилось
                if (newAccess === null || newAccess === undefined) {
                    this.fileService.revokeGroupAccess(fileId, groupId).subscribe();
                } else {
                    this.fileService.shareFileWithGroup(fileId, groupId, newAccess).subscribe();
                }
            }
        }
    
        this.visible = false;
    }

    cancel() {
        this.visible = false;
    }
}
