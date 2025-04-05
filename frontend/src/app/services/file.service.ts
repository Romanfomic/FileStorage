import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environment';
import { FileMetadata } from '../interfaces/fileData';

@Injectable({ providedIn: 'root' })
export class FileService {
    private http = inject(HttpClient);

    private baseUrl = `${environment.apiUrl}/api/files`;

    getUserFiles(): Observable<FileMetadata[]> {
        return this.http.get<FileMetadata[]>(this.baseUrl);
    }
}
