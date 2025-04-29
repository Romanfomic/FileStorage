import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environment';
import { FileMetadata } from '../interfaces/fileData';
import { FileVersion } from '../interfaces/version';

@Injectable({ providedIn: 'root' })
export class FileService {
    private http = inject(HttpClient);

    private baseUrl = `${environment.apiUrl}/api/files`;
    private sharedUrl = `${environment.apiUrl}/api/shared-files`;
    private versionUrl = `${environment.apiUrl}/api/versions`;

    getUserFiles(): Observable<FileMetadata[]> {
        return this.http.get<FileMetadata[]>(this.baseUrl);
    }

    uploadFile(formData: FormData): Observable<any> {
        return this.http.post(`${this.baseUrl}/upload`, formData, {
            responseType: 'text',
        });
    }

    deleteFile(fileId: number): Observable<any> {
        return this.http.delete(`${this.baseUrl}/${fileId}`);
    }    

    downloadFile(fileId: number): Observable<Blob> {
        return this.http.get(`${this.baseUrl}/${fileId}`, {
            responseType: 'blob'
        });
    }

    renameFile(fileId: number, name: string): Observable<any> {
        return this.http.put(`${this.baseUrl}/${fileId}`, { name });
    }      

    getSharedFiles(): Observable<FileMetadata[]> {
        return this.http.get<FileMetadata[]>(`${this.sharedUrl}`);
    }

    shareFileWithUser(fileId: number, userId: number, accessId: number): Observable<any> {
        return this.http.post(`${this.baseUrl}/${fileId}/share/user`, {
            user_id: userId,
            access_id: accessId
        });
    }
    
    shareFileWithGroup(fileId: number, groupId: number, accessId: number): Observable<any> {
        return this.http.post(`${this.baseUrl}/${fileId}/share/group`, {
            group_id: groupId,
            access_id: accessId
        });
    }
    
    revokeUserAccess(fileId: number, userId: number): Observable<any> {
        return this.http.delete(`${this.baseUrl}/${fileId}/share/user/${userId}`);
    }
    
    revokeGroupAccess(fileId: number, groupId: number): Observable<any> {
        return this.http.delete(`${this.baseUrl}/${fileId}/share/group/${groupId}`);
    }
    
    getFilePermissions(fileId: number): Observable<any> {
        return this.http.get(`${this.baseUrl}/${fileId}/permissions`);
    }

    getFileVersions(fileId: number): Observable<FileVersion[]> {
        return this.http.get<FileVersion[]>(`${this.baseUrl}/${fileId}/versions`);
    }

    createVersion(fileId: number, name: string): Observable<any> {
        return this.http.post(`${this.baseUrl}/${fileId}/version`, { name });
    }

    setCurrentVersion(fileId: number, versionId: number): Observable<any> {
        return this.http.put(`${this.baseUrl}/${fileId}/version`, { version_id: versionId });
    }

    deleteVersion(versionId: number): Observable<any> {
        return this.http.delete(`${this.versionUrl}/${versionId}`);
    }

    renameVersion(versionId: number, newName: string): Observable<any> {
        return this.http.put(`${this.versionUrl}/${versionId}`, { name: newName });
    }
}
