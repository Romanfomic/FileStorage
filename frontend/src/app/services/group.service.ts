import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../environment';
import { Group } from '../interfaces/group';

@Injectable({ providedIn: 'root' })
export class GroupService {
    private http = inject(HttpClient);
    private baseUrl = `${environment.apiUrl}/api/groups`;

    getGroups(): Observable<Group[]> {
        return this.http.get<Group[]>(this.baseUrl);
    }

    getGroupById(id: number): Observable<Group> {
        return this.http.get<Group>(`${this.baseUrl}/${id}`);
    }

    createGroup(group: Partial<Group>): Observable<Group> {
        return this.http.post<Group>(this.baseUrl, group);
    }

    updateGroup(id: number, group: Partial<Group>): Observable<void> {
        return this.http.put<void>(`${this.baseUrl}/${id}`, group);
    }

    deleteGroup(id: number): Observable<void> {
        return this.http.delete<void>(`${this.baseUrl}/${id}`);
    }
}
