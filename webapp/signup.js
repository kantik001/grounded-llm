const form = document.getElementById('signupForm');
const planSelect = document.getElementById('plan');
const result = document.getElementById('result');
const errorEl = document.getElementById('error');
const submitBtn = document.getElementById('submitBtn');

async function loadPlans() {
  const res = await fetch('/api/v1/plans');
  const data = await res.json();
  if (!data.success) throw new Error(data.error || 'Failed to load plans');
  planSelect.innerHTML = '';
  (data.plans || []).forEach((p) => {
    if (p.contact_sales) return;
    const opt = document.createElement('option');
    opt.value = p.id;
    const price = p.price_monthly === 0 ? 'Free' : `$${p.price_monthly}/mo`;
    opt.textContent = `${p.label} — ${price}`;
    planSelect.appendChild(opt);
  });
}

form.addEventListener('submit', async (e) => {
  e.preventDefault();
  errorEl.hidden = true;
  submitBtn.disabled = true;
  try {
    const body = {
      org_name: document.getElementById('orgName').value.trim(),
      email: document.getElementById('email').value.trim(),
      plan: planSelect.value,
    };
    const res = await fetch('/api/v1/signup', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    const data = await res.json();
    if (!res.ok || !data.success) {
      throw new Error(data.error || 'Signup failed');
    }
    form.hidden = true;
    result.hidden = false;
    document.getElementById('tenantId').textContent = data.tenant_id;
    const chat = document.getElementById('chatLink');
    chat.href = data.chat_url || `/?tenant_id=${data.tenant_id}`;
  } catch (err) {
    errorEl.textContent = err.message || String(err);
    errorEl.hidden = false;
  } finally {
    submitBtn.disabled = false;
  }
});

loadPlans().catch((err) => {
  errorEl.textContent = err.message || String(err);
  errorEl.hidden = false;
});
