import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../environment';
import { Permission } from '../interfaces/permission';

@Injectable({ providedIn: 'root' })
export class PermissionService {
    private http = inject(HttpClient);
    private baseUrl = `${environment.apiUrl}/api/permissions`;

    getPermissions(): Observable<Permission[]> {
        return this.http.get<Permission[]>(this.baseUrl);
    }

    getPermission(id: number): Observable<Permission> {
        return this.http.get<Permission>(`${this.baseUrl}/${id}`);
    }
}
