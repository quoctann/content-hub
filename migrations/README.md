Postgres naming convention

Schemas:
- Use domain schemas: auth, content, billing, audit.
- Use shared for reusable custom functions/triggers.
- Keep public for extensions only unless there is a good reason.

General:
- Always schema-qualify shared functions and extension functions in migrations.
- Do not rely on search_path in migrations or function bodies.
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