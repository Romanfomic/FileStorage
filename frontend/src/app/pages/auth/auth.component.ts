import { Component, inject } from '@angular/core';
import { FormBuilder, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { AuthService } from '../../services/auth.service'; 
import { Router } from '@angular/router';
import { NgIf } from '@angular/common';

@Component({
    selector: 'app-auth',
    imports: [
        NgIf,
        FormsModule, 
        ReactiveFormsModule,
    ],
    templateUrl: './auth.component.html',
    styleUrl: './auth.component.less',
    standalone: true,
})
export class AuthComponent {
    private authService = inject(AuthService);
    private router = inject(Router);
    private fb = inject(FormBuilder);

    loginForm = this.fb.group({
        login: this.fb.nonNullable.control<string>('', [Validators.required]),
        password: this.fb.nonNullable.control<string>('', [Validators.required]),
    });
    
    errorMessage: string = '';
  
    onSubmit(): void {
        if (this.loginForm.invalid) return;
        const { login, password } = this.loginForm.value

        this.authService.login({login, password}).subscribe({
            next: (res) => {
                localStorage.setItem('token', res.token);
                this.router.navigate(['/dashboard']);
            },
            error: () => {
                this.errorMessage = 'Неверный логин или пароль';
            }
        });
    }
}
