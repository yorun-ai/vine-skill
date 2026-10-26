// Runs only in the temporary generated example; credentials are demo-only.
const assert = require('node:assert/strict');
const { createVrpcClient, getClientInstanceId } = require('@yorun-ai/vrpc');
const { createAuthApiService, createGreetingApiService, createWelcomeApiService } = require('./compiled/index.js');
const [business, admin] = process.argv.slice(2);
const clientInfo = { clientName: 'greeting.smoke', clientVersion: '1.0.0', clientInstanceId: getClientInstanceId() };
const client = createVrpcClient({ prefixUrl: `${business}/api/invoke`, clientInfo });
const auth = createAuthApiService(client);
const greetings = createGreetingApiService(client);
const welcome = createWelcomeApiService(client);

async function main() {
  let login;
  for (let attempt = 0; ; attempt++) {
    try {
      login = await auth.login({ username: 'admin', password: 'admin123' });
      break;
    } catch (error) {
      if (error.code !== 'SERVICE_UNAVAILABLE' || attempt === 99) throw error;
      await new Promise(resolve => setTimeout(resolve, 100));
    }
  }
  assert.equal(login.success, true);
  assert.ok(login.token);
  await assert.rejects(welcome.getWelcome(null), error => error.code === 'UNAUTHORIZED');
  await assert.rejects(welcome.getWelcome(null, { headers: { Authorization: 'token invalid' } }),
    error => error.code === 'UNAUTHORIZED');
  const result = await welcome.getWelcome(null, { headers: { Authorization: `token ${login.token}` } });
  assert.equal(result.username, 'admin');
  const created = await greetings.createGreeting({ message: 'generated TypeScript client' });
  const page = await greetings.listGreetings({ pageToken: null, pageSize: null });
  assert.ok(page.items.some(item => item.id === created.id));
  await assert.rejects(greetings.createGreeting({ message: '' }), error => error.code === 'VALIDATION_FAILED');

  // This validation fixture checks the direct Hub Admin listener at the wire boundary.
  const adminClient = createVrpcClient({ prefixUrl: `${admin}/api/invoke`, clientInfo, traceMode: 'direct' });
  const readOnly = await adminClient.invoke({ serviceName: 'vine.hub.admin.AdminApiService', methodName: 'readOnly', params: null });
  assert.equal(readOnly, false);
  console.log('Generated TypeScript: login, Actor admission, CRUD/errors; direct Hub Admin: passed.');
}
main().catch(error => { console.error(error); process.exitCode = 1; });
