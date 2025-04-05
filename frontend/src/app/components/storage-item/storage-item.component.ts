import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FileMetadata } from '../../interfaces/fileData';

@Component({
    selector: 'app-storage-item',
    standalone: true,
    imports: [CommonModule],
    templateUrl: './storage-item.component.html',
    styleUrl: './storage-item.component.less',
})
export class StorageItemComponent {
    @Input() file!: FileMetadata;

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
}
