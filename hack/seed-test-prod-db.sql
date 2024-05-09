CREATE TABLE IF NOT EXISTS public.users (
  id TEXT NOT NULL DEFAULT gen_random_uuid(),
  age int NOT NULL,
  current_salary float NOT NULL,
  first_name TEXT NOT NULL,
  last_name TEXT NOT null,

  CONSTRAINT users_pk PRIMARY KEY (id)
);

-- CREATE TABLE IF NOT EXISTS public.user_pets (
--   id TEXT NOT NULL DEFAULT gen_random_uuid(),
--   user_id TEXT NOT NULL,

--   CONSTRAINT user_pets_pk PRIMARY KEY (id),
--   CONSTRAINT fk_user_pets_users_id FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE cascade
-- );
