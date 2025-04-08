import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environment';

@Injectable({
    providedIn: 'root'
})
export class AuthService {
    private http = inject(HttpClient);
    private baseUrl = `${environment.apiUrl}/login`;

    login(data: { login?: string; password?: string }): Observable<{ token: string }> {
        return this.http.post<{ token: string }>(`${this.baseUrl}`, data);
    }

    logout(): void {
        localStorage.removeItem('token');
    }

    isAuthenticated(): boolean {
        return !!localStorage.getItem('token');
    }
}
