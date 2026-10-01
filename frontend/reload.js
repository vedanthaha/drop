const { Client } = require('pg');
const c = new Client({connectionString: 'postgresql://postgres:vedant233445566@db.jwnksdmbaexaicjxlxrv.supabase.co:5432/postgres'});
c.connect().then(() => c.query("NOTIFY pgrst, 'reload schema';")).then(() => console.log('reloaded')).catch(console.error).finally(()=>c.end());
