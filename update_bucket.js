const https = require('https');

const data = JSON.stringify({ 
  file_size_limit: 5368709120 // 5GB limit
});

const options = {
  hostname: 'jwnksdmbaexaicjxlxrv.supabase.co',
  path: '/storage/v1/bucket/media',
  method: 'PUT',
  headers: {
    'apikey': 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imp3bmtzZG1iYWV4YWljanhseHJ2Iiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc5MDc4MDc0NywiZXhwIjoyMTA2MzU2NzQ3fQ.IqA8MRfVKVGrZuiTRuwetJxEgZrCWzPJ-XKOCAg3G20',
    'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imp3bmtzZG1iYWV4YWljanhseHJ2Iiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc5MDc4MDc0NywiZXhwIjoyMTA2MzU2NzQ3fQ.IqA8MRfVKVGrZuiTRuwetJxEgZrCWzPJ-XKOCAg3G20',
    'Content-Type': 'application/json',
    'Content-Length': data.length
  }
};

const req = https.request(options, (res) => {
  let body = '';
  res.on('data', d => body += d);
  res.on('end', () => console.log('Response:', body));
});
req.on('error', e => console.error('Error:', e));
req.write(data);
req.end();
