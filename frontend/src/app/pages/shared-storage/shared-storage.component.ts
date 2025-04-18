import { Component, inject } from '@angular/core';
import { CommonModule, NgFor, AsyncPipe } from '@angular/common';
import { FileService } from '../../services/file.service';
import { FileMetadata } from '../../interfaces/fileData';
import { DialogModule } from 'primeng/dialog';
import { UserService } from '../../services/user.service';
import { User } from '../../interfaces/user';
import { EMPTY, tap, catchError, Observable } from 'rxjs';
import { StorageItemComponent } from '../../components/storage-item/storage-item.component';

@Component({
    selector: 'app-shared-storage',
    imports: [
        CommonModule,
        NgFor,
        AsyncPipe,
        DialogModule,
        StorageItemComponent,
    ],
    templateUrl: './shared-storage.component.html',
    styleUrl: './shared-storage.component.less'
})
export class SharedStorageComponent {
    private fileService = inject(FileService);
    private userService = inject(UserService);

    load$ = this.fileService.getSharedFiles().pipe(
        tap(files => this.files = files)
    );

    files: FileMetadata[] = [];
    selectedFile: FileMetadata | null = null;
    fileOwner: User | null = null;

    showDialog = false;
    showInfoDialog = false;
    contextMenuPosition = { x: '0px', y: '0px' };

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
}
