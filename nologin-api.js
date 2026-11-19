/* Experimental adapter for NoLogin's public editor storage. No SDK required.
 * Only use pages you created for this app. This is not an official NoLogin API.
 */
(function (root) {
  'use strict';
  const base = 'https://firestore.googleapis.com/v1/projects/no-login-044/databases/(default)/documents/documents';
  function page(id) {
    if (!/^torbellino-[a-z0-9-]{8,100}$/.test(id)) throw new Error('Use a unique torbellino- prefixed ID.');
    return `${base}/${encodeURIComponent(id)}`;
  }
  function encode(value) {
    if (typeof value === 'string') return {stringValue:value};
    if (typeof value === 'boolean') return {booleanValue:value};
    if (typeof value === 'number' && Number.isSafeInteger(value)) return {integerValue:String(value)};
    if (Array.isArray(value)) return {arrayValue:{values:value.map(encode)}};
    if (value && typeof value === 'object') return {mapValue:{fields:fields(value)}};
    throw new Error('Unsupported field type');
  }
  function fields(value) { return Object.fromEntries(Object.entries(value).map(([k,v])=>[k,encode(v)])); }
  function decode(value) {
    if ('stringValue' in value) return value.stringValue;
    if ('integerValue' in value) return Number(value.integerValue);
    if ('booleanValue' in value) return value.booleanValue;
    if ('arrayValue' in value) return (value.arrayValue.values || []).map(decode);
    if ('mapValue' in value) return Object.fromEntries(Object.entries(value.mapValue.fields || {}).map(([k,v])=>[k,decode(v)]));
    return null;
  }
  async function request(url, method='GET', body) {
    const response = await fetch(url, {method, credentials:'omit', cache:'no-store',
      signal:AbortSignal.timeout(20000),
      ...(body ? {headers:{'Content-Type':'application/json'},body:JSON.stringify(body)} : {})});
    const data = await response.json();
    if (!response.ok) {
      const error = new Error(data.error?.message || `HTTP ${response.status}`);
      error.status=response.status; error.code=data.error?.status; throw error;
    }
    return data;
  }
  function unpack(doc) {
    const data=Object.fromEntries(Object.entries(doc.fields || {}).map(([k,v])=>[k,decode(v)]));
    // Deliberately do not expose or implement password/lock bypasses.
    if (data.passwordSet || data.isLocked || data.isTimeLocked) throw new Error('Protected pages are unsupported.');
    if (data.expirationTimestamp && data.expirationTimestamp <= Date.now()) throw new Error('Page expired.');
    return {id:doc.name.split('/').pop(),content:data.textTabs?.[0]?.content ?? data.text ?? '',
      version:doc.updateTime, expiresAt:data.expirationTimestamp, tabs:data.textTabs || []};
  }
  async function read(id) { return unpack(await request(page(id))); }
  async function create(id, content, hours=1) {
    page(id);
    if (!Number.isFinite(hours) || hours<=0 || hours>24) throw new Error('Expiry must be between 0 and 24 hours.');
    const now=Date.now();
    return unpack(await request(`${base}?documentId=${encodeURIComponent(id)}`, 'POST', {fields:fields({
      text:String(content), textTabs:[{id:'torbellino-main',title:'Text 1',content:String(content)}],
      files:[],passwordSet:false,password:'',createdAt:now,expirationTimestamp:now+Math.round(hours*3600000)
    })}));
  }
  async function update(id, content, version) {
    if (!version) throw new Error('Pass the version returned by read() to prevent lost updates.');
    const current=await read(id);
    if (current.version !== version) { const e=new Error('Page changed; read it again and merge.'); e.code='FAILED_PRECONDITION'; throw e; }
    const tabs=current.tabs.length ? current.tabs.map((t,i)=>i ? t : {...t,content:String(content)})
      : [{id:'torbellino-main',title:'Text 1',content:String(content)}];
    const params=new URLSearchParams({'currentDocument.updateTime':version});
    params.append('updateMask.fieldPaths','text'); params.append('updateMask.fieldPaths','textTabs');
    return unpack(await request(`${page(id)}?${params}`, 'PATCH', {fields:fields({text:String(content),textTabs:tabs})}));
  }
  root.NoLoginAPI=Object.freeze({create,read,update});
})(globalThis);
