const { Client } = require('pg');
const c = new Client({connectionString: 'postgresql://postgres:vedant233445566@db.jwnksdmbaexaicjxlxrv.supabase.co:5432/postgres'});
c.connect()
 .then(() => c.query("SELECT id, selected_format FROM download_jobs WHERE platform='instagram' ORDER BY created_at DESC LIMIT 1;"))
 .then(res => {
   const val = res.rows[0].selected_format;
   console.log(val, Buffer.from(val));
 })
 .catch(console.error)
 .finally(()=>c.end());
