const { createClient } = require('@supabase/supabase-js');
const sb = createClient('https://jwnksdmbaexaicjxlxrv.supabase.co', 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imp3bmtzZG1iYWV4YWljanhseHJ2Iiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc5MDc4MDc0NywiZXhwIjoyMTA2MzU2NzQ3fQ.IqA8MRfVKVGrZuiTRuwetJxEgZrCWzPJ-XKOCAg3G20');

async function check() {
  const { data, error } = await sb.storage.getBucket('media');
  console.log('Bucket check:', data, error);
  if (error) {
    console.log('Attempting to create bucket...');
    const { data: cData, error: cError } = await sb.storage.createBucket('media', { public: false });
    console.log('Create result:', cData, cError);
  }
}
check().catch(console.error);
