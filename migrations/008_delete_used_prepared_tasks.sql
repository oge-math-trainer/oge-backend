BEGIN;

DELETE FROM prepared_tasks
WHERE used = true;

COMMIT;
