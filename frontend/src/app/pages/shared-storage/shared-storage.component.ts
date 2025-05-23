import { Component, inject } from '@angular/core';
import { CommonModule, NgFor, AsyncPipe } from '@angular/common';
import { FileService } from '../../services/file.service';
import { FileMetadata } from '../../interfaces/fileData';
import { DialogModule } from 'primeng/dialog';
import { UserService } from '../../services/user.service';
import { User } from '../../interfaces/user';
import { EMPTY, tap, catchError, Observable } from 'rxjs';
import { StorageItemComponent } from '../../components/storage-item/storage-item.component';
import { FormsModule } from '@angular/forms';
import { InputTextModule } from 'primeng/inputtext';

@Component({
    selector: 'app-shared-storage',
    imports: [
        CommonModule,
        NgFor,
        AsyncPipe,
        DialogModule,
        StorageItemComponent,
        FormsModule,
        InputTextModule,
    ],
    templateUrl: './shared-storage.component.html',
    styleUrl: './shared-storage.component.less'
})
export class SharedStorageComponent {
    private fileService = inject(FileService);
    private userService = inject(UserService);

    searchQuery = '';

    files: FileMetadata[] = [];
    selectedFile: FileMetadata | null = null;
    fileOwner: User | null = null;

    showDialog = false;
    showInfoDialog = false;
    showPreviewDialog = false;
    contextMenuPosition = { x: '0px', y: '0px' };

    previewUrl: string | null = null;
    previewType: 'image' | 'video' | 'audio' | null = null;

    load$ = this.fileService.getSharedFiles().pipe(
        tap(files => this.files = files)
    );
    action$!: Observable<any>;

    onRightClick(event: MouseEvent, file: FileMetadata) {
        event.preventDefault();
        this.selectedFile = file;
        this.contextMenuPosition = {
            x: `${event.clientX}px`,
            y: `${event.clientY}px`
        };
        this.showDialog = true;
    }

    onSearchChange(query: string) {
        this.loadFiles(query);
    }

    loadFiles(search = '') {
        this.load$ = this.fileService.getSharedFiles(search).pipe(
            tap((files) => {
                if (!files) this.files = []
                else this.files = files
            })
        );
    }

    downloadFile(file: FileMetadata) {
        this.fileService.downloadFile(file.file_id).pipe(
            tap((blob) => {
                const link = document.createElement('a');
                link.href = window.URL.createObjectURL(blob);
                link.download = file.name;
                link.click();
                window.URL.revokeObjectURL(link.href);
            }),
            catchError((err) => {
                console.error(err);
                return EMPTY;
            })
        ).subscribe(() => this.showDialog = false);
    }

    viewInfo(file: FileMetadata) {
        this.selectedFile = file;
        this.showDialog = false;
        this.showInfoDialog = true;

        this.userService.getUserById(file.owner_id!).pipe(
            tap(user => this.fileOwner = user),
            catchError(err => {
                console.error(err);
                return EMPTY;
            })
        ).subscribe();
    }

    previewFile(file: FileMetadata) {
        this.action$ = this.fileService.downloadFile(file.file_id).pipe(
            tap((blob) => {
                const url = URL.createObjectURL(blob);

                const ext = file.name.split('.').pop()?.toLowerCase();
                if (ext) {
                    if (['png', 'jpg', 'jpeg', 'gif', 'webp'].includes(ext)) {
                        this.previewType = 'image';
                    } else if (['mp4', 'webm'].includes(ext)) {
                        this.previewType = 'video';
                    } else if (['mp3', 'wav', 'ogg'].includes(ext)) {
                        this.previewType = 'audio';
                    } else {
                        this.previewType = null;
                    }
                }
    
                if (this.previewType) {
                    this.previewUrl = url;
                    this.showPreviewDialog = true;
                } else {
                    // download if cannot show preview
                    const link = document.createElement('a');
                    link.href = url;
                    link.download = file.name;
                    link.click();
                    URL.revokeObjectURL(url);
                }
            }),
            catchError((error) => {
                console.error('Preview error', error);
                return EMPTY;
            }),
        )
    }

    closePreview() {
        this.showPreviewDialog = false;
        if (this.previewUrl) {
            URL.revokeObjectURL(this.previewUrl);
        }
        this.previewUrl = null;
        this.previewType = null;
    }
}
