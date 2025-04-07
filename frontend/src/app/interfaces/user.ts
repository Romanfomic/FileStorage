export interface User {
    user_id: number;
    mail: string;
    login: string;
    name: string;
    surname: string;
    type: string;
    role_id: number | null;
    group_id: number | null;
}
