import { AsyncPipe, DatePipe, NgFor, NgIf } from '@angular/common';
import { Component, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ButtonModule } from 'primeng/button';
import { DialogModule } from 'primeng/dialog';
import { InputTextModule } from 'primeng/inputtext';
import { RadioButtonModule } from 'primeng/radiobutton';
import { catchError, EMPTY, Observable, tap } from 'rxjs';
import { FileMetadata } from '../../../interfaces/fileData';
import { FileService } from '../../../services/file.service';
import { FileVersion } from '../../../interfaces/version';

@Component({
    selector: 'app-file-version-dialog',
    imports: [
        RadioButtonModule,
        DialogModule,
        ButtonModule,
        InputTextModule,
        NgFor,
        NgIf,
        FormsModule,
        DatePipe,
        AsyncPipe,
    ],
    templateUrl: './file-version-dialog.component.html',
    styleUrl: './file-version-dialog.component.less'
})
export class FileVersionDialogComponent {
    private fileService = inject(FileService);

    action$!: Observable<any>;

    file: FileMetadata | null = null;
    showDialog = false;

    versions: FileVersion[] = [];
    currentVersionId: number | null = null;
    newVersionName = '';
    editingVersionId: number | null = null;
    editingVersionName = '';

    open(file: FileMetadata) {
        this.file = file;
        this.showDialog = true;
        this.loadVersions(file.file_id);
    }

    close() {
        this.showDialog = false;
    }

    loadVersions(fileId: number) {
        this.action$ = this.fileService.getFileVersions(fileId)
            .pipe(
                tap((versions: FileVersion[]) => {
                    this.versions = versions;
                    const current = versions.find(v => v.is_current);
                    this.currentVersionId = current?.version_id ?? null;
                }),
                    catchError((err) => {
                    console.error('Load versions error', err);
                    return EMPTY;
                })
            );
    }

    setCurrentVersion(version: FileVersion) {
        this.action$ = this.fileService.setCurrentVersion(this.file!.file_id, version.version_id)
            .pipe(
                tap(() => this.loadVersions(this.file!.file_id)),
                catchError((err) => {
                    console.error('Set current version error', err);
                    return EMPTY;
                })
            );
    }

    createVersion() {
        if (!this.newVersionName.trim()) return;

        this.action$ = this.fileService.createVersion(this.file!.file_id, this.newVersionName.trim())
            .pipe(
                tap(() => {
                    this.newVersionName = '';
                    this.loadVersions(this.file!.file_id);
                }),
                catchError((err) => {
                    console.error('Create version error', err);
                    return EMPTY;
                })
            );
    }

    deleteVersion(version: FileVersion) {
        console.log(version);
        if (version.version_id === this.currentVersionId) return;

        this.action$ = this.fileService.deleteVersion(version.version_id)
            .pipe(
                tap(() => this.loadVersions(this.file!.file_id)),
                catchError((err) => {
                    console.error('Delete version error', err);
                    return EMPTY;
                })
            )
    }

    enableEditing(version: FileVersion) {
        this.editingVersionId = version.version_id;
        this.editingVersionName = version.name;
    }

    saveEdit(version: FileVersion) {
        if (!this.editingVersionName.trim()) return;

        this.action$ = this.fileService.renameVersion(version.version_id, this.editingVersionName.trim())
            .pipe(
                tap(() => {
                    this.editingVersionId = null;
                    this.loadVersions(this.file!.file_id);
                }),
                catchError((err) => {
                    console.error('Rename version error', err);
                    return EMPTY;
                })
            );
    }

    cancelEdit() {
        this.editingVersionId = null;
        this.editingVersionName = '';
    }
}
