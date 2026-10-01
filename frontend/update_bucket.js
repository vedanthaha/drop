const { createClient } = require('@supabase/supabase-js');
const sb = createClient('https://jwnksdmbaexaicjxlxrv.supabase.co', 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imp3bmtzZG1iYWV4YWljanhseHJ2Iiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc5MDc4MDc0NywiZXhwIjoyMTA2MzU2NzQ3fQ.IqA8MRfVKVGrZuiTRuwetJxEgZrCWzPJ-XKOCAg3G20');

async function update() {
  const { data, error } = await sb.storage.updateBucket('media', { public: false, fileSizeLimit: 524288000 });
  console.log('Update result:', data, error);
}
update().catch(console.error);
