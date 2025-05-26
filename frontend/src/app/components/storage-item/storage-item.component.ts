import { Component, EventEmitter, inject, Input, Output } from '@angular/core';
import { AsyncPipe, CommonModule, DatePipe, NgIf } from '@angular/common';
import { FileMetadata } from '../../interfaces/fileData';
import { Observable, tap } from 'rxjs';
import { UserService } from '../../services/user.service';
import { User } from '../../interfaces/user';
import { Group } from '../../interfaces/group';
import { GroupService } from '../../services/group.service';

@Component({
    selector: 'app-storage-item',
    standalone: true,
    imports: [CommonModule, DatePipe, NgIf, AsyncPipe],
    templateUrl: './storage-item.component.html',
    styleUrl: './storage-item.component.less',
})
export class StorageItemComponent {
    private userService = inject(UserService);
    private groupService = inject(GroupService);

    @Input() file!: FileMetadata;

    @Output() preview = new EventEmitter<FileMetadata>();

    user?: User;
    group?: Group;

    loadUser$?: Observable<any>;
    
    loadGroup$?: Observable<any>;

    ngOnInit(): void {
        if (!!this.file.access_id) {
            this.loadUser$ = this.userService.getUserById(this.file.owner_id!).pipe(
                tap((user) => this.user = user)
            );
    
            if (this.file.group_ids) {
                this.loadGroup$ = this.groupService.getGroupById(this.file.group_ids[0]).pipe(
                    tap((group) => this.group = group)
                );
            }
        }
    }

    get isFolder(): boolean {
        return this.file.type === 'folder';
    }

    get iconSrc(): string {
        if (this.isFolder) return '/assets/icons/folder.png';
        const ext = this.file.name.split('.').pop()?.toLowerCase();
        switch (ext) {
            case 'pdf': return '/assets/icons/pdf.png';
            case 'doc':
            case 'docx': return '/assets/icons/doc.png';
            case 'csv': return '/assets/icons/csv.png';
            case 'mp3': return '/assets/icons/audio.png';
            case 'mp4': return '/assets/icons/video.png';
            case 'jpeg':
            case 'jpg':
            case 'png': return '/assets/icons/image.png';
            case 'txt': return '/assets/icons/file.png';
            case 'js':
            case 'ts':
            case 'json': return '/assets/icons/code.png';
            default: return '/assets/icons/file.png';
        }
    }

    getAccess(): string {
        if (!this.file.access_id) return ''
        
        if (this.file.access_id === 1) return 'чтение'
        return 'редактирование'
    }

    getAuthor(): string {
        if (this.group) {
            return this.group.name;
        }

        return (this.user?.surname ?? '') + ' ' + (this.user?.name ?? '');
    }
}
