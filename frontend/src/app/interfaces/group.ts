export interface Group {
    group_id: number;
    name: string;
    description: string;
    parent_id?: number | null;
    depth: number;
    children?: Group[];
}