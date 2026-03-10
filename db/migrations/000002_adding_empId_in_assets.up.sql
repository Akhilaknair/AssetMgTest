ALTER TABLE assets
    ADD COLUMN IF NOT EXISTS current_assigned_to uuid NULL;

ALTER TABLE assets
    ADD CONSTRAINT fk_assets_current_user
        FOREIGN KEY (current_assigned_to)
            REFERENCES users(id);


