// Standalone disposable smoke-test helper; reads ephemeral credentials from stdin.
let raw=''; for await (const chunk of process.stdin) raw+=chunk;
const input=JSON.parse(raw);
async function check(token, expected){
 await new Promise((resolve,reject)=>{
  const socket=new WebSocket(input.endpoint,'graphql-ws');
  const timer=setTimeout(()=>{socket.close();reject(new Error('subscription timeout'));},10000);
  const fail=()=>{clearTimeout(timer);socket.close();reject(new Error('subscription failed'));};
  socket.addEventListener('open',()=>socket.send(JSON.stringify({type:'connection_init',payload:{headers:{Authorization:'Bearer '+token}}})));
  socket.addEventListener('message',(event)=>{
   const m=JSON.parse(event.data);
   if(m.type==='connection_ack') socket.send(JSON.stringify({id:'test',type:'start',payload:{query:'subscription { notifications { id read } }'}}));
   if(m.type==='error'||m.type==='connection_error') fail();
   if(m.type==='data'){
    const records=m.payload.data?.notifications;
    if(m.payload.errors||!Array.isArray(records)||(expected?!records.some(r=>r.id===expected):records.length!==0)){fail();return;}
    clearTimeout(timer);socket.send(JSON.stringify({id:'test',type:'stop'}));socket.close();resolve();
   }
  });
  socket.addEventListener('error',fail);
 });
}
await check(input.alice,input.notification); await check(input.bob,null);
process.stdout.write('GraphQL WebSocket notification reads and user isolation: passed\n');
