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

    userAccessMap: { [userId: number]: boolean } = {};
    userAccessLevelMap: { [userId: number]: number } = {};
    groupAccessMap: { [groupId: number]: boolean } = {};
    groupAccessLevelMap: { [groupId: number]: number } = {};

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

        this.loadUsers$ = this.userService.getAllUsers().pipe((
            tap((users) => this.allUsers = users)
        ))

        this.loadGroups$ = this.groupService.getGroups().pipe((
            tap((groups) => this.allGroups = groups)
        ))
        
        this.load$ = this.fileService.getFilePermissions(this.file.file_id).pipe(
            tap((res) => {
                if (!!res.users)
                    for (let u of res.users) {
                        this.userAccessMap[u.user_id] = true;
                        this.userAccessLevelMap[u.user_id] = u.access_id;
                    }

                if (!!res.groups)
                    for (let g of res.groups) {
                        this.groupAccessMap[g.group_id] = true;
                        this.groupAccessLevelMap[g.group_id] = g.access_id;
                    }
            }),
        )
    }    

    save() {
        const fileId = this.file.file_id;

        for (let userId in this.userAccessMap) {
            const id = Number(userId);
            if (this.userAccessMap[id]) {
                this.fileService.shareFileWithUser(fileId, id, this.userAccessLevelMap[id]).subscribe();
            } else {
                this.fileService.revokeUserAccess(fileId, id).subscribe();
            }
        }

        for (let groupId in this.groupAccessMap) {
            const id = Number(groupId);
            if (this.groupAccessMap[id]) {
                this.fileService.shareFileWithGroup(fileId, id, this.groupAccessLevelMap[id]).subscribe();
            } else {
                this.fileService.revokeGroupAccess(fileId, id).subscribe();
            }
        }

        this.visible = false;
    }

    cancel() {
        this.visible = false;
    }
}
