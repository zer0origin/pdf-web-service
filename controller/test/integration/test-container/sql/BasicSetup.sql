create table if not exists document_table
(
    "Document_UUID"   uuid not null
        constraint "Document_Table_pk"
            primary key,
    "Document_Base64" text not null
);

create table if not exists documentmeta_table
(
    "Document_UUID"   uuid not null
        constraint documentmeta_table_pk
            primary key
        constraint documentmeta_table_document_table_null_fk
            references document_table
            on delete cascade,
    "Number_Of_Pages" integer,
    "Height"          numeric,
    "Width"           numeric,
    "Images"          json
);

create table if not exists selection_table
(
    "Selection_UUID"   uuid not null
        constraint selection_table_pk
            primary key,
    "Document_UUID"    uuid
        constraint "selection_table_document_table_Document_UUID_fk"
            references document_table
            on delete cascade,
    "isCompleted"      boolean default false,
    "Settings"         json,
    "Selection_bounds" json,
    "Page_Words"       json
);