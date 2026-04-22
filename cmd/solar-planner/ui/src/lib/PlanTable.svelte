<script>
	import {
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell,
		Badge
	} from 'flowbite-svelte';
	import PlanInspector from './PlanInspector.svelte';
	import MonthlyBreakdown from './MonthlyBreakdown.svelte';

	let { results, plans } = $props();
	let selectedPlanName = $state(null);

	// Derived state: Calculate yearly total and sort by cheapest first
	const sortedResults = $derived.by(() => {
		if (!results) return [];
		
		return results
			.map((plan) => {
				const totalCents = plan.Data.reduce((acc, month) => acc + month.Cents, 0);
				const monthCount = plan.Data.length || 1;
				const projectedYearly = (totalCents / monthCount) * 12;

				return {
					...plan,
					totalCents,
					projectedYearly,
					isCredit: projectedYearly < 0
				};
			})
			.sort((a, b) => a.projectedYearly - b.projectedYearly);
	});

	function togglePlan(name) {
		selectedPlanName = selectedPlanName === name ? null : name;
	}

	function formatCurrency(cents) {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(cents / 100);
	}
</script>

<div class="relative overflow-x-auto shadow-md sm:rounded-lg">
	<Table hoverable={true}>
		<TableHead>
			<TableHeadCell>Rank</TableHeadCell>
			<TableHeadCell>Plan Name</TableHeadCell>
			<TableHeadCell>Projected Annual Cost</TableHeadCell>
			<TableHeadCell>Avg Monthly</TableHeadCell>
			<TableHeadCell>Status</TableHeadCell>
		</TableHead>
		<TableBody>
			{#each sortedResults as plan, i (plan.Name)}
				<TableBodyRow
					class="cursor-pointer"
					onclick={() => togglePlan(plan.Name)}
				>
					<TableBodyCell class="font-medium">
						{i + 1}
					</TableBodyCell>
					<TableBodyCell>
						{plan.Name}
					</TableBodyCell>
					<TableBodyCell class={plan.isCredit ? 'font-bold text-green-600' : 'text-gray-900'}>
						{formatCurrency(plan.projectedYearly)}
					</TableBodyCell>
					<TableBodyCell>
						{formatCurrency(plan.projectedYearly / 12)}
					</TableBodyCell>
					<TableBodyCell>
						{#if plan.isCredit}
							<Badge color="green">Net Exporter</Badge>
						{:else if plan.projectedYearly < 50000}
							<Badge color="blue">Efficient</Badge>
						{:else}
							<Badge color="red">High Cost</Badge>
						{/if}
					</TableBodyCell>
				</TableBodyRow>
				{#if selectedPlanName === plan.Name}
					<TableBodyRow class="bg-gray-50 dark:bg-gray-800/50">
						<TableBodyCell colspan="5" class="p-6">
							<PlanInspector plan={plans.find((p) => p.Name === plan.Name)} />

							<div class="border-t pt-6 dark:border-gray-700">
								<h4 class="mb-4 text-sm font-bold tracking-widest text-gray-500 uppercase">
									Monthly Breakdown
								</h4>
								<MonthlyBreakdown monthlyData={plan.Data} />
							</div>
						</TableBodyCell>
					</TableBodyRow>
				{/if}
			{/each}
		</TableBody>
	</Table>
</div>

<style>
	/* Add any specific transitions here */
</style>
