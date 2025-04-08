import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../environment';
import { Role } from '../interfaces/role';

@Injectable({ providedIn: 'root' })
export class RoleService {
    private http = inject(HttpClient);
    private baseUrl = `${environment.apiUrl}/api/roles`;

    getRoles(): Observable<Role[]> {
        return this.http.get<Role[]>(this.baseUrl);
    }

    getRoleById(id: number): Observable<Role> {
        return this.http.get<Role>(`${this.baseUrl}/${id}`);
    }

    createRole(role: Partial<Role>): Observable<Role> {
        return this.http.post<Role>(this.baseUrl, role);
    }

    updateRole(id: number, role: Partial<Role>): Observable<void> {
        return this.http.put<void>(`${this.baseUrl}/${id}`, role);
    }

    deleteRole(id: number): Observable<void> {
        return this.http.delete<void>(this.baseUrl);
    }
}
