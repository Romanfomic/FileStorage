export interface FileMetadata {
    file_id: number;
    name: string;
    type: string;
    full_path: string;
    create_date: string;
    edit_date: string;
    version_id: number;
    group_ids?: number[];
    owner_id?: number;
    access_id?: number;
}