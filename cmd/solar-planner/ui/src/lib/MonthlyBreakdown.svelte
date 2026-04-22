<script>
	import {
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';
	let { monthlyData } = $props();

	const format = (cents) =>
		(cents / 100).toLocaleString('en-US', { style: 'currency', currency: 'USD' });
</script>

<div class="rounded-xl bg-gray-50 p-4 dark:bg-gray-900">
	<Table striped={true}>
		<TableHead>
			<TableHeadCell>Month</TableHeadCell>
			<TableHeadCell>Import (kWh)</TableHeadCell>
			<TableHeadCell>Export (kWh)</TableHeadCell>
			<TableHeadCell>Energy Net</TableHeadCell>
			<TableHeadCell>TDU Charges</TableHeadCell>
			<TableHeadCell>Bill</TableHeadCell>
		</TableHead>
		<TableBody>
			{#each monthlyData as m}
				<TableBodyRow>
					<TableBodyCell>{m.Month}</TableBodyCell>
					<TableBodyCell>{m.MBI.Import.toFixed(1)}</TableBodyCell>
					<TableBodyCell>{m.MBI.Export.toFixed(1)}</TableBodyCell>
					<TableBodyCell>{format(m.MBI.EnergyChargeCents - m.MBI.SolarCreditCents)}</TableBodyCell>
					<TableBodyCell>{format(m.MBI.TDUChargeCents)}</TableBodyCell>
					<TableBodyCell class="font-bold">{format(m.Cents)}</TableBodyCell>
				</TableBodyRow>
			{/each}
		</TableBody>
	</Table>
</div>
