<script>
	import { Card, Badge } from 'flowbite-svelte';
	import { InfoCircleOutline } from 'flowbite-svelte-icons';

	let { plan } = $props();
	console.log(plan);
</script>

{#if plan && plan.Export}
	<Card size="xl" padding="md" class="border-l-4 border-l-blue-500">
		<div class="mb-4 flex items-start justify-between">
			<h3 class="text-xl font-bold text-gray-900 dark:text-white">{plan.Name} Configuration</h3>
			<Badge color="indigo" class="uppercase">{plan.Export.Model}</Badge>
		</div>

		<div class="grid grid-cols-1 gap-6 text-sm md:grid-cols-2">
			<div>
				<h4 class="mb-2 font-semibold text-gray-700 dark:text-gray-300">Billing Logic</h4>
				<ul class="space-y-2">
					<li class="flex items-center gap-2">
						<span class="h-2 w-2 rounded-full bg-blue-500"></span>
						Netting Mode: <strong>{plan.Netting.Mode}</strong>
					</li>
					{#if plan.Netting.NoNetExport}
						<li class="font-medium text-amber-600">
							⚠️ Energy credits cannot reduce the bill below $0 (plus fees).
						</li>
					{/if}
					{#if plan.Netting.CapToImport}
						<li class="font-medium text-amber-600">
							⚠️ Export credit is limited to your total import volume.
						</li>
					{/if}
				</ul>
			</div>

			<div>
				<h4 class="mb-2 font-semibold text-gray-700 dark:text-gray-300">Rates (Cents)</h4>
				<ul class="space-y-2">
					<li>Energy Import: <strong>{plan.Charges.ImportCentsPerKWh}¢</strong></li>
					<li>TDU Import: <strong>{plan.Charges.TDUCentsPerKWh}¢</strong></li>
					<li>
						Fixed Monthly Fees: <strong
							>{(plan.Charges.BaseCents + plan.Charges.TDUBaseCents) / 100}</strong
						>
					</li>
					<li>
						Buyback:
						<strong>
							{plan.Export.Model === 'fixed' ? `${plan.Export.FixedRate}¢` : 'Wholesale (RTW)'}
						</strong>
					</li>
				</ul>
			</div>
		</div>

		{#if plan.TOU && plan.TOU.Enabled}
			<div
				class="mt-4 rounded border border-blue-100 bg-blue-50 p-3 dark:border-blue-800 dark:bg-blue-900/20"
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
