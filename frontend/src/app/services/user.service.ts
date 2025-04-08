import { HttpClient } from "@angular/common/http";
import { inject, Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { environment } from "../../environment";
import { User } from "../interfaces/user";

@Injectable({ providedIn: 'root' })
export class UserService {
    private http = inject(HttpClient);
    private baseUrl = `${environment.apiUrl}/api/users`;

    getUserById(id: number): Observable<User> {
        return this.http.get<User>(`${this.baseUrl}/${id}`);
    }

    getAllUsers(): Observable<User[]> {
        return this.http.get<User[]>(this.baseUrl);
    }

    createUser(user: Partial<User>): Observable<User> {
        return this.http.post<User>(this.baseUrl, user);
    }

    updateUser(user: User): Observable<User> {
        return this.http.put<User>(`${this.baseUrl}/${user.user_id}`, user);
    }

    deleteUser(id: number): Observable<void> {
        return this.http.delete<void>(`${this.baseUrl}/${id}`);
    }
}
