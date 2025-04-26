import { Component, inject, ViewChild } from '@angular/core';
import { NgFor, AsyncPipe } from '@angular/common';
import { FileService } from '../../services/file.service';
import { CommonModule } from '@angular/common';
import { catchError, EMPTY, finalize, Observable, tap } from 'rxjs';
import { FileMetadata } from '../../interfaces/fileData';
import { StorageItemComponent } from '../../components/storage-item/storage-item.component';
import { DialogModule } from 'primeng/dialog';
import { UserService } from '../../services/user.service';
import { User } from '../../interfaces/user';
import { FileUploadModule } from 'primeng/fileupload';
import { FileAccessDialogComponent } from "../../components/dialogs/file-access-dialog/file-access-dialog.component";
import { FormsModule } from '@angular/forms';
import { ButtonModule } from 'primeng/button';
import { InputTextModule } from 'primeng/inputtext';

@Component({
    selector: 'app-storage',
    standalone: true,
    imports: [
    NgFor,
    AsyncPipe,
    CommonModule,
    StorageItemComponent,
    DialogModule,
    FileUploadModule,
    FileAccessDialogComponent,
    FormsModule,
    ButtonModule,
    InputTextModule
],
    templateUrl: './storage.component.html',
    styleUrl: './storage.component.less',
})
export class StorageComponent {
    private fileService = inject(FileService);
    private userService = inject(UserService);

    @ViewChild(FileAccessDialogComponent) accessDialog!: FileAccessDialogComponent;

    load$ = this.fileService.getUserFiles().pipe(
        tap((files) => {
            this.files = files;
        })
    );

    action$!: Observable<any>;

    files: FileMetadata[] = [];
    selectedFile: FileMetadata | null = null;

    fileOwner: User | null = null;

    showDialog = false;
    showInfoDialog = false;
    showPreviewDialog = false;
    showRenameDialog = false;

    newFileName = '';
    contextMenuPosition = { x: '0px', y: '0px' };
    previewUrl: string | null = null;
    previewType: 'image' | 'video' | 'audio' | null = null;

    onFileSelected(event: { files: File[] }) {
        const file = event.files?.[0];
        if (!file) return;

        const formData = new FormData();
        formData.append('file', file);

        this.action$ = this.fileService.uploadFile(formData).pipe(
            tap((value) => {
                console.log(value);
                this.refreshFiles();
            }),
            catchError((error) => {
                console.error('Upload error', error);
                return EMPTY;
            })
        );
    }
    
    refreshFiles() {
        this.load$ = this.fileService.getUserFiles().pipe(
            tap(files => this.files = files),
        );
    }

    onRightClick(event: MouseEvent, file: FileMetadata) {
        event.preventDefault();
        this.selectedFile = file;
        this.contextMenuPosition = {
            x: `${event.clientX}px`,
            y: `${event.clientY}px`
        };
        this.showDialog = true;
    }

    downloadFile(file: FileMetadata) {
        this.action$ = this.fileService.downloadFile(file.file_id).pipe(
            tap((blob) => {
                const link = document.createElement('a');
                link.href = window.URL.createObjectURL(blob);
                link.download = file.name;
                link.click();
                window.URL.revokeObjectURL(link.href);
            }),
            catchError((error) => {
                console.error('Download error', error);

                return EMPTY;
            }),
            finalize(() => this.showDialog = false)
        )
    }

    viewInfo(file: FileMetadata) {
        this.selectedFile = file;
        this.showDialog = false;
        this.showInfoDialog = true;

        this.action$ = this.userService.getUserById(file.owner_id!).pipe(
            tap((user) => this.fileOwner = user),
            catchError((error) => {
                console.error('Get user error', error)
                return EMPTY;
            }),
        );
    }

    openRenameDialog(file: FileMetadata) {
        this.selectedFile = file;
        this.newFileName = file.name;
        this.showDialog = false;
        this.showRenameDialog = true;
    }

    openShareDialog(file: FileMetadata) {
        this.showDialog = false;
        this.accessDialog.open(file);
    }
    
    deleteFile(file: FileMetadata) {
        this.action$ = this.fileService.deleteFile(file.file_id).pipe(
            tap(() => {
                this.showDialog = false;
                this.refreshFiles();
            })
        );
    }

    renameFile() {
        if (!this.selectedFile) return;
    
        this.action$ = this.fileService.renameFile(this.selectedFile.file_id, this.newFileName).pipe(
            tap(() => {
                this.refreshFiles();
                this.showRenameDialog = false;
            }),
            catchError((error) => {
                console.error('Rename error', error);
                return EMPTY;
            })
        );
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
