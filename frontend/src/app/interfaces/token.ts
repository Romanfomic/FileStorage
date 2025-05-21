export interface TokenPayload {
    user_id: number;
    type: string;
    permissions: string[];
    exp: number;
    iat: number;
}