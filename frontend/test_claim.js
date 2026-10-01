const url = 'https://jwnksdmbaexaicjxlxrv.supabase.co/rest/v1/rpc/claim_next_download_job';
const key = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imp3bmtzZG1iYWV4YWljanhseHJ2Iiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc5MDc4MDc0NywiZXhwIjoyMTA2MzU2NzQ3fQ.IqA8MRfVKVGrZuiTRuwetJxEgZrCWzPJ-XKOCAg3G20';

fetch(url, {
  method: 'POST',
  headers: {
    'apikey': key,
    'Authorization': `Bearer ${key}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({ p_worker_id: 'test', p_lease_seconds: 900 })
})
.then(res => res.text())
.then(console.log)
.catch(console.error);
