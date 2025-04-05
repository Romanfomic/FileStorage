import { Component, inject } from '@angular/core';
import { NgFor, AsyncPipe } from '@angular/common';
import { FileService } from '../../services/file.service';
import { CommonModule } from '@angular/common';
import { tap } from 'rxjs';
import { FileMetadata } from '../../interfaces/fileData';
import { StorageItemComponent } from '../../components/storage-item/storage-item.component';

@Component({
    selector: 'app-storage',
    standalone: true,
    imports: [NgFor, AsyncPipe, CommonModule, StorageItemComponent],
    templateUrl: './storage.component.html',
    styleUrl: './storage.component.less',
})
export class StorageComponent {
    private fileService = inject(FileService);

    load$ = this.fileService.getUserFiles().pipe(
        tap((files) => {
            this.files = files;
        })
    );

    files: FileMetadata[] = [];
}
