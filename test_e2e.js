const fs = require('fs');

const URLS = [
  "https://in.pinterest.com/pin/615867317839642992/"
];

async function run() {
  console.log("Starting E2E Tests...\n");
  for (const url of URLS) {
    console.log(`Testing URL: ${url}`);
    
    // 1. Resolve
    const resResolve = await fetch('http://localhost:3002/api/resolve', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url })
    });
    
    if (!resResolve.ok) {
      console.log(`  -> Resolve FAILED: ${resResolve.status}`);
      continue;
    }
    const resolveData = await resResolve.json();
    console.log(`  -> Resolved: ${resolveData.title} (${resolveData.formats.length} formats)`);
    
    if (resolveData.formats.length === 0) {
      console.log(`  -> No formats available.`);
      continue;
    }

    // Pick the smallest video format to bypass Supabase 50MB free-tier limit
    const formatId = "upscale-4x";
    console.log(`  -> Selected Format ID: ${formatId}`);

    // 2. Download Job
    const resDownload = await fetch('http://localhost:3002/api/download', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ 
        mediaId: resolveData.id, 
        formatId: formatId,
        output: "jpg" 
      })
    });
    
    if (!resDownload.ok) {
      console.log(`  -> Job creation FAILED: ${resDownload.status} ${await resDownload.text()}`);
      continue;
    }
    const jobData = await resDownload.json();
    console.log(`  -> Job created: ${jobData.jobId}`);
    
    // 3. Poll Job Status
    let attempts = 0;
    let jobReady = false;
    while (attempts < 60) {
      const resJob = await fetch(`http://localhost:3002/api/jobs/${jobData.jobId}`);
      if (!resJob.ok) break;
      const statusData = await resJob.json();
      
      if (statusData.status === 'ready') {
        jobReady = true;
        console.log(`  -> Job READY! Storage Path: ${statusData.filename}`);
        break;
      } else if (statusData.status === 'error') {
        console.log(`  -> Job ERROR: ${statusData.error_message}`);
        break;
      }
      
      await new Promise(r => setTimeout(r, 2000));
      attempts++;
    }
    
    if (!jobReady) {
      console.log(`  -> Job did not complete in time.`);
      continue;
    }
    
    // 4. Get Signed URL
    const resFiles = await fetch(`http://localhost:3002/api/files/${jobData.jobId}`);
    if (!resFiles.ok) {
      console.log(`  -> Signed URL fetch FAILED: ${resFiles.status}`);
      continue;
    }
    const fileData = await resFiles.json();
    console.log(`  -> Signed URL retrieved successfully.`);
    console.log(`----------------------------------------\n`);
  }
}

run().catch(console.error);
