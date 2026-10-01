const fs = require('fs');
const { Client } = require('pg');

async function run() {
  const sql = fs.readFileSync(process.argv[2], 'utf-8');
  const client = new Client({
    connectionString: process.env.DATABASE_URL
  });
  await client.connect();
  await client.query(sql);
  await client.end();
}
run().catch(console.error);
