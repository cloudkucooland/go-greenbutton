<script>
	import { Alert, List, Li, Card, Button } from 'flowbite-svelte';
	import {
		ExclamationCircleSolid,
		ChevronDownOutline,
		ChevronUpOutline
	} from 'flowbite-svelte-icons';

	let isOpen = $state(false);
</script>

<Card size="none" class="border-amber-200 bg-amber-50 p-4 dark:bg-amber-900/10">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-2 text-amber-800 dark:text-amber-400">
			<ExclamationCircleSolid size="md" />
			<h3 class="text-lg font-bold">Alpha Version: Manual verification of data is required</h3>
		</div>
		<Button
			pill={true}
			color="light"
			size="xs"
			aria-label={isOpen ? 'Collapse warnings' : 'Expand warnings'}
			class="border-none bg-transparent hover:bg-amber-100 dark:hover:bg-amber-800/50"
			onclick={() => (isOpen = !isOpen)}
		>
			{#if isOpen}
				<ChevronUpOutline size="sm" />
			{:else}
				<ChevronDownOutline size="sm" />
			{/if}
		</Button>
	</div>

	{#if isOpen}
		<div class="mt-4 space-y-6 transition-all duration-300">
			<p class="text-sm text-gray-700 dark:text-gray-300">
				This simulator is a community-driven tool in <strong>Alpha</strong>. While the math is based
				on published EFL (Electricity Facts Label) data, utility rates change frequently.
			</p>

			<div class="grid grid-cols-1 gap-8 text-sm leading-relaxed md:grid-cols-2">
				<div>
					<h4 class="mb-2 font-bold text-gray-900 underline decoration-amber-500 dark:text-white">
						Current Assumptions
					</h4>
					<List tag="ul" class="space-y-1 text-sm text-gray-600 dark:text-gray-400">
						<Li>TDU delivery charges are estimated based on Oncor/Centerpoint 2026 rates.</Li>
						<Li
							>Wholesale (RTW) buyback is estimated using monthly-weighted historical averages
							(e.g., March: 0.56¢, Aug: 2.43¢).</Li
						>
						<Li>Calculations assume 15-minute interval data from Smart Meter Texas.</Li>
						<Li>No battery storage or load shifting is currently simulated.</Li>
					</List>
				</div>

				<div>
					<h4 class="mb-2 font-bold text-gray-900 underline decoration-amber-500 dark:text-white">
						Critical Warnings
					</h4>
					<List tag="ul" class="space-y-1 text-sm text-gray-600 dark:text-gray-400">
						<Li
							>This tool models what you would have been charged in the past were you on different
							plans.</Li
						>
						<Li>This is <strong>not</strong> a financial guarantee of future bills.</Li>
						<Li>Taxes and municipal franchise fees are not included in projections.</Li>
						<Li>Always verify the current EFL with the provider before switching plans.</Li>
					</List>
				</div>
			</div>

			<div class="space-y-3">
				<Alert color="yellow" class="border border-amber-300">
					<span class="font-medium">Found a bug?</span>
					If these numbers don't match your actual bill, please open an issue on GitHub with your anonymized
					CSV data.
				</Alert>
				<Alert color="yellow" class="border border-amber-300">
					<span class="font-medium">Plan missing or has incorrect data?</span>
					If plans are not up-to-date, please open an issue on GitHub with the correct data.
				</Alert>
			</div>
		</div>
	{/if}
</Card>
