Postgres naming convention

Schemas:
- Application tables live in the schema set by DATABASE_SCHEMA (prod: meme).
  The app and the migrator both send it as search_path and create it if missing,
  so migrations must NOT qualify table/index/trigger targets (write `content`,
  not `meme.content`). Renaming the schema is then a config change, not a
  migration edit.
- Keep public for extensions and shared functions/triggers.

General:
- Always schema-qualify shared functions and extension functions (public.unaccent,
  public.trg_update_updated_at). Built-in SQL constructs are never qualified
  (coalesce, not public.coalesce).
- Function bodies must pin their own search_path (SET search_path = public).
- Never edit a migration that has already been applied anywhere; add a new one.
  Every .down.sql must actually run: CI applies up -> down -all -> up.
- Use snake_case only.
- Avoid generic names such as handle, process, update, uuid.

Functions:
- Normal function: <verb>_<object>[_<qualifier>]
- Trigger function: trg_<action>
- Examples: shared.gen_uuid(), shared.trg_set_updated_at()

Triggers:
- trg_<table>_<action>_<timing>
- Timing suffix: bi, bu, bd, ai, au, ad
- Examples: trg_users_set_updated_at_bu, trg_posts_generate_slug_bi

Indexes:
- Normal: idx_<table>__<columns>
- Unique: uidx_<table>__<columns>
- GIN: gin_<table>__<columns>
- BRIN: brin_<table>__<columns>
- Partial: idx_<table>__<columns>__where_<condition>
- Expression: idx_<table>__expr_<description>

Constraints:
- Primary key: pk_<table>
- Foreign key: fk_<from_table>__<column>__<to_table>
- Unique: uq_<table>__<columns>
- Check: ck_<table>__<rule>