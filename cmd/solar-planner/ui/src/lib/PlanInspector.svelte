{#if plan && plan.Export}
	<Card size="xl" padding="md" class="border-l-4 border-l-blue-500">
		<div class="mb-4 flex items-start justify-between">
			<div>
				<h3 class="text-xl font-bold text-gray-900 dark:text-white">
					{plan.Name}
				</h3>
				{#if plan.Provider}
					<p class="text-sm text-gray-500">
						by
						{#if plan.ProviderURL}
							<a href={plan.ProviderURL} target="_blank" class="text-blue-600 underline"
								>{plan.Provider}</a
							>
						{:else}
							{plan.Provider}
						{/if}
					</p>
				{/if}
			</div>
			<div class="flex flex-col items-end gap-2">
				<Badge color="indigo" class="uppercase">{plan.Export.Model}</Badge>
				{#if plan.PlanURL}
					<a href={plan.PlanURL} target="_blank" class="text-xs text-blue-500 underline"
						>View EFL Label</a
					>
				{/if}
			</div>
		</div>

		{#if plan.Details}
			<p class="mb-6 text-sm text-gray-600 italic dark:text-gray-400">
				"{plan.Details}"
			</p>
		{/if}

		<div class="grid grid-cols-1 gap-8 text-sm md:grid-cols-3">
			<div>
				<h4 class="mb-2 font-bold text-gray-900 uppercase dark:text-white">Billing Logic</h4>
				<ul class="space-y-2">
					<li class="flex items-center gap-2">
						<span class="h-2 w-2 rounded-full bg-blue-500"></span>
						Netting: <strong>{plan.Netting.Mode || 'Monthly'}</strong>
					</li>
					{#if plan.Netting.NoNetExport}
						<li class="font-medium text-amber-600">
							⚠️ Credits cannot reduce bill below $0.
						</li>
					{/if}
					{#if plan.Netting.CapToImport}
						<li class="font-medium text-amber-600">
							⚠️ Export credit capped at total import.
						</li>
					{/if}
					{#if plan.Credits && plan.Credits.Enabled}
						<li class="text-green-600">
							✅ Credit rollover enabled
							{#if plan.Credits.ExpirationMonths > 0}
								({plan.Credits.ExpirationMonths}mo)
							{/if}
						</li>
					{/if}
				</ul>
			</div>

			<div>
				<h4 class="mb-2 font-bold text-gray-900 uppercase dark:text-white">Rates (Cents)</h4>
				<ul class="space-y-2">
					<li>Energy: <strong>{plan.Charges.ImportCentsPerKWh}¢</strong></li>
					<li>TDU: <strong>{plan.Charges.TDUCentsPerKWh}¢</strong></li>
					<li>
						Monthly Fees: <strong
							>{formatCurrency(plan.Charges.BaseCents + plan.Charges.TDUBaseCents)}</strong
						>
					</li>
					<li>
						Buyback:
						<strong class="text-blue-600">
							{plan.Export.Model === 'fixed'
								? `${plan.Export.FixedRate}¢`
								: plan.Export.Model === 'netting'
									? `${plan.Export.FixedRate}¢ (1:1)`
									: 'Wholesale (RTW)'}
						</strong>
					</li>
				</ul>
			</div>

			<div>
				<h4 class="mb-2 font-bold text-gray-900 uppercase dark:text-white">Requirements</h4>
				<ul class="space-y-2">
					{#if plan.Battery && plan.Battery.Required}
						<li class="font-semibold text-red-500">🔋 Battery Required</li>
					{:else}
						<li class="text-gray-500">No battery required</li>
					{/if}
					{#if plan.Limits && (plan.Limits.MaxKWhPerMonth > 0 || plan.Limits.MaxCreditDollars > 0)}
						<li class="text-amber-600">
							🛑 Limits:
							{#if plan.Limits.MaxKWhPerMonth > 0}
								{plan.Limits.MaxKWhPerMonth}kWh/mo
							{/if}
							{#if plan.Limits.MaxCreditDollars > 0}
								${plan.Limits.MaxCreditDollars}/mo
							{/if}
						</li>
					{/if}
				</ul>
			</div>
		</div>

		{#if plan.TOU && plan.TOU.Enabled}
			<div
				class="mt-6 rounded border border-blue-100 bg-blue-50 p-3 dark:border-blue-800 dark:bg-blue-900/20"
			>
				<p class="flex items-center gap-2 text-blue-800 dark:text-blue-300">
					<InfoCircleOutline size="sm" />
					Time-of-Use active with {plan.TOU.Periods.length} custom periods.
				</p>
			</div>
		{/if}
	</Card>
{:else}
	<div class="p-4 text-gray-500 italic">Loading plan details...</div>
{/if}

<script>
	import { Card, Badge } from 'flowbite-svelte';
	import { InfoCircleOutline } from 'flowbite-svelte-icons';

	let { plan } = $props();

	function formatCurrency(cents) {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(cents / 100);
	}
</script>
