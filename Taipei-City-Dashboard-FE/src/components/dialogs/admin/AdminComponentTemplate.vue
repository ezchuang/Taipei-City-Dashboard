<!-- Developed by Taipei Urban Intelligence Center 2023-2024-->

<!-- THIS COMPONENT IS UNDER DEVELOPMENT -->
<!-- NOT PRODUCTION READY -->

<script setup>
	import { ref } from "vue";
	import { useAdminStore } from "../../../store/adminStore";
	import { useDialogStore } from "../../../store/dialogStore";
	import http from "../../../router/axios";
	import InputTags from "../../utilities/forms/InputTags.vue";
	import DialogContainer from "../DialogContainer.vue";

	const dialogStore = useDialogStore();
	const adminStore = useAdminStore();

	const default_params_component = {
		index: "",
		name: "",
		history_config: "",
		map_config_indexes: [],
		map_filter: "",
		time_from: null,
		time_to: null,
		update_freq: null,
		update_freq_unit: "minute",
		short_desc: "",
		long_desc: "",
		use_case: "",
		query_type: "",
		query_charts: [
			{
				source: "",
				links: [],
				contributors: [],
				query_chart: "",
				city: "taipei",
			},
		],
	};

	const default_params_chart = {
		colors: [],
		types: [],
		unit: "",
	};

	const default_params_maps = [
		{
			title: "",
			type: "",
			source: "",
			size: "",
			icon: "",
			paint: "",
			property: "",
		},
	];

	const default_params_table = {
		table: {
			name: "",
			columns: [
				{
					is_primary_key: false,
					name: "",
					type: "text",
					notNull: false,
					comment: "",
				},
			],
		},

		file: null,
	};

	const params_component = ref({ ...default_params_component });
	const params_chart = ref({ ...default_params_chart });
	const params_maps = ref([...default_params_maps]);
	const params_table = ref({ ...default_params_table });

	const options_column_types = [
		{ label: "char", value: "char" },
		{ label: "bigint", value: "bigint" },
		{ label: "boolean", value: "boolean" },
		{ label: "character", value: "character" },
		{ label: "date", value: "date" },
		{ label: "double precision", value: "double" },
		{ label: "json", value: "json" },
		{ label: "text", value: "text" },
		{ label: "char[]", value: "char_list" },
		{ label: "bigint[]", value: "bigint_list" },
		{ label: "boolean[]", value: "boolean_list" },
		{ label: "character[]", value: "character_list" },
		{ label: "date[]", value: "date_list" },
		{ label: "double precision[]", value: "double_list" },
		{ label: "json[]", value: "json_list" },
		{ label: "text[]", value: "text_list" },
	];

	const options_query_type = [
		{ label: "2D", value: "two_d" },
		{ label: "3D", value: "three_d" },
		{ label: "time", value: "time" },
		{ label: "percent", valueent: "percent" },
		{ label: "map_legend", valuelegend: "map_legend" },
	];

	const options_city = ref([
		{ label: "臺北市", value: "taipei" },
		{ label: "雙北", value: "metrotaipei" },
	]);

	const options_button_tab = ref([
		{ label: "組件", order: 0 },
		{ label: "地圖圖層", order: 1 },
		{ label: "資料表", order: 2 },
	]);

	const tempInputStorage = ref({
		link: "",
		contributor: "",
		chartColor: "#000000",
		chartType: "",
		historyColor: "#000000",
	});

	const currentTab = ref(0);

	function addColumn() {
		params_table.value.table.columns.push({
			is_primary_key: false,
			name: "",
			type: "text",
			description: "",
			not_null: false,
			comment: "",
		});
	}

	function addMapConfig() {
		params_maps.value.push({
			title: "",
			type: "",
			source: "",
			size: "",
			icon: "",
			paint: "",
			property: "",
		});
	}

	function addQueryChart() {
		if (params_component.value.query_charts.length >= options_city.value.length) return;
		params_component.value.query_charts.push({
			source: "",
			links: [],
			contributors: [],
			query_chart: "",
			city: "",
		});
	}

	function handleClose() {
		dialogStore.hideAllDialogs();
		params_component.value = { ...default_params_component };
	}

	async function handleSubmit() {
		const res_validation = isValidComponentParams();

		if (!res_validation.result) {
			dialogStore.showNotification("fail", res_validation.message);
			return;
		}

		try {
			const components = {
				index: params_component.value.index,
				name: params_component.value.name,
			};
			const component_charts = {
				index: params_component.value.index,
				color: params_chart.value.colors,
				types: params_chart.value.types,
				unit: params_chart.value.unit,
			};
			const component_maps = params_maps.value.map((map) => {
				return Object.assign({}, map, { index: params_component.value.index });
			});

			const query_charts = params_component.value.query_charts.map((query_chart) => {
				return Object.assign({}, query_chart, {
					index: params_component.value.index,
					history_config: params_component.value.history_config,
					map_filter: params_component.value.map_filter,
					time_from: params_component.value.time_from,
					time_to: params_component.value.time_to,
					update_frep: params_component.value.update_freq,
					update_freq_unit: params_component.value.update_freq_unit,
					short_desc: params_component.value.short_desc,
					long_desc: params_component.value.long_desc,
					use_case: params_component.value.use_case,
					query_type: params_component.value.query_type,
				});
			});

			const body = {
				components,
				component_maps,
				component_charts,
				query_charts,
			};

			const createRes = await http.post(`/component/`, params_component.value);
			const getRes = await http.get(`/component/${createRes.data.data.id}`);
			adminStore.getComponentData(getRes.data.data);
			dialogStore.showNotification("success", "新建組建成功");
			dialogStore.hideAllDialogs();
			dialogStore.showDialog("adminComponentSettings");
		} catch (error) {
			console.error(error);
		}
	}

	async function handleTableSubmit() {
		const res_validation = isValidTableParams();

		if (!res_validation.result) {
			dialogStore.showNotification("fail", res_validation.message);
			return;
		}

		const formData = new FormData();

		formData.append("file", params_table.value.file);
		formData.append("tableName", params_table.value.table.name);
		formData.append("columns", params_table.value.table.columns);

		try {
			const response = await http.post("/api/v1/tables/import", formData, {
				headers: {
					"Content-Type": "multipart/form-data",
				},
			});
			dialogStore.showNotification("success", "上傳成功");
		} catch (error) {
			console.error(error);
			dialogStore.showNotification("fail", "上傳失敗");
		}
	}

	function handleFileChange(event) {
		params_table.value.file = event.target.files[0];
	}

	function isValidComponentParams() {
		const { index, name, update_freq, update_freq_unit, short_desc, long_desc, use_case } = params_component.value;
		if (
			!index ||
			!name ||
			(!update_freq && update_freq !== 0) ||
			!update_freq_unit ||
			!short_desc ||
			!long_desc ||
			!use_case
		) {
			return { result: false, message: "新建失敗，請確實填寫所有必填欄位" };
		}

		const { colors, types, unit } = params_chart.value;
		if (colors.length === 0 || types.length === 0 || !unit.trim()) {
			return { result: false, message: "新建失敗，請確實填寫所有圖表必填欄位" };
		}

		for (const query_chart of params_component.value.query_charts) {
			if (!query_chart.source.trim() || !query_chart.query_chart.trim() || !query_chart.city.trim()) {
				return { result: false, message: "新建失敗，請確實填寫 query chart 必填欄位" };
			}
		}
		const json_string = [];
		if (params_component.value.history_config.trim())
			json_string.push({
				label: "history_config",
				str: params_component.value.history_config,
			});

		if (params_component.value.map_filter)
			json_string.push({
				label: "map_filter",
				str: params_component.value.map_filter,
			});

		for (const map of params_maps.value) {
			const { title, type, paint, property } = map;
			if (!title || !type) return { result: false, message: "新建失敗，請確實填寫地圖圖層所有必填欄位" };

			if (paint.trim())
				json_string.push({
					label: "paint",
					str: paint,
				});

			if (property.trim())
				json_string.push({
					label: "property",
					str: property,
				});
		}

		for (const str of json_string) {
			if (!isValidJSON(str.str)) return { result: false, message: `無效的JSON格式 - 欄位:${str.label}` };
		}

		return { result: true, message: "" };
	}

	const tables = await http.get("/tables");
	function isValidTableParams() {
		if (!params_table.value.file) return { result: false, message: "請選擇檔案" };
		if (!params_table.value.table) return { result: false, message: "請填寫資料表名稱" };

		if (tables.data.indexOf(params_table.value.table.name) === -1)
			return { result: false, message: "資料表已經存在" };

		const map = {};
		for (let column of params_table.value.table.columns) {
			if (Object.hasOwn(map, column.name)) return { result: false, message: "重複的表格名稱" };
			else map[column.name] = column.name;

			if (!column.name.trim() || !column.type) return { result: false, message: "請填寫完整表格資訊" };
		}

		return { result: true, message: "" };
	}

	function isValidJSON(str) {
		try {
			JSON.parse(str);
			return true;
		} catch (error) {
			return false;
		}
	}

	function isShowTimeToBlock(time_to) {
		switch (time_to) {
			case "now":
			case "static":
			case "current":
				return false;

			default:
				return true;
		}
	}
</script>

<template>
	<DialogContainer :dialog="`adminAddComponentTemplate`" @on-close="handleClose">
		<div class="admincomponenttemplate">
			<div class="admincomponenttemplate-header">
				<h2>組件設定</h2>
				<button @click="handleSubmit">確定新增</button>
			</div>
			<div class="admincomponenttemplate-tabs">
				<template v-for="tab in options_button_tab" :key="tab.order">
					<button :class="currentTab === tab.order ? 'active' : ''" @click="currentTab = tab.order">
						{{ tab.label }}
					</button>
				</template>
			</div>
			<div class="admincomponenttemplate-content">
				<div class="admincomponenttemplate-settings">
					<div v-if="currentTab === 0" class="admincomponenttemplate-settings-items">
						<label>組件名稱* ({{ params_component.name.length }}/10)</label>
						<input v-model="params_component.name" type="text" :minlength="1" :maxlength="10" required />
						<div class="two-block">
							<label>組件 Index*</label>
						</div>
						<div class="two-block">
							<input v-model="params_component.index" type="text" required />
						</div>
						<label>顏色*</label>
						<InputTags
							:tags="params_chart.colors"
							@deletetag="
								(index) => {
									params_chart.colors.splice(index, 1);
								}
							"
							@updatetagorder="
								(updatedTags) => {
									params_chart.colors = updatedTags;
								}
							"
						/>
						<input
							v-model="tempInputStorage.chartColor"
							type="text"
							:minlength="1"
							@keypress.enter="
								() => {
									params_chart.colors.push(tempInputStorage.chartColor);
									tempInputStorage.chartColor = '';
								}
							"
						/>

						<label>圖表類型*</label>
						<InputTags
							:tags="params_chart.types"
							@deletetag="
								(index) => {
									params_chart.types.splice(index, 1);
								}
							"
							@updatetagorder="
								(updatedTags) => {
									params_chart.types = updatedTags;
								}
							"
						/>
						<input
							v-model="tempInputStorage.chartType"
							type="text"
							:minlength="1"
							@keypress.enter="
								() => {
									params_chart.types.push(tempInputStorage.chartType);
									tempInputStorage.chartType = '';
								}
							"
						/>
						<label>單位*</label>
						<input v-model="params_chart.unit" type="text" :minlength="1" :maxlength="12" required />
						<label>歷史設定(JSON)</label>
						<textarea v-model="params_component.history_config" />
						<label>地圖Filter(JSON)</label>
						<textarea v-model="params_component.map_filter" />
						<label>資料區間</label>
						<div class="time-block">
							<select v-model="params_component.time_from" required>
								<option value="day_ago">一天前</option>
								<option value="week_ago">一週前</option>
								<option value="month_ago">一個月前</option>
								<option value="quarter_ago">一季前</option>
								<option value="halfyear_ago">半年前</option>
								<option value="year_ago">一年前</option>
								<option value="twoyear_ago">兩年前</option>
								<option value="fiveyear_ago">五年前</option>
								<option value="tenyear_ago">十年前</option>
								<option value="now">現在</option>
								<option value="static">固定資料</option>
								<option value="current">即時資料</option>
							</select>
							<span v-show="isShowTimeToBlock(params_component.time_from)">～</span>
							<select
								v-show="isShowTimeToBlock(params_component.time_from)"
								v-model="params_component.time_to"
								required
							>
								<option value="day_ago">一天前</option>
								<option value="week_ago">一週前</option>
								<option value="month_ago">一個月前</option>
								<option value="quarter_ago">一季前</option>
								<option value="halfyear_ago">半年前</option>
								<option value="year_ago">一年前</option>
								<option value="twoyear_ago">兩年前</option>
								<option value="fiveyear_ago">五年前</option>
								<option value="tenyear_ago">十年前</option>
								<option value="now">現在</option>
								<option value="static">固定資料</option>
								<option value="current">即時資料</option>
							</select>
						</div>
						<label>更新頻率* (0 = 不定期更新)</label>
						<div class="two-block">
							<input v-model="params_component.update_freq" type="number" :min="0" :max="31" required />
							<select v-model="params_component.update_freq_unit">
								<option value="minute">分</option>
								<option value="hour">時</option>
								<option value="day">天</option>
								<option value="week">週</option>
								<option value="month">月</option>
								<option value="year">年</option>
							</select>
						</div>
						<label required>組件簡述* ({{ params_component.short_desc.length }}/50)</label>
						<textarea v-model="params_component.short_desc" :minlength="1" :maxlength="50" required />
						<label>組件詳述* ({{ params_component.long_desc.length }}/100)</label>
						<textarea v-model="params_component.long_desc" :minlength="1" :maxlength="100" required />
						<label>範例情境* ({{ params_component.use_case.length }}/100)</label>
						<textarea v-model="params_component.use_case" :minlength="1" :maxlength="100" required />

						<label for="query_type">圖表類型</label>
						<select v-model="params_component.query_type" name="query_type" id="query_type">
							<template v-for="query_type in options_query_type" :key="query_type.value">
								<option :value="query_type.value">{{ query_type.label }}</option>
							</template>
						</select>
						<div class="flex mt-2 justify-between">
							<label>多城市組件設定 (query chart)</label>
							<button @click.prevent="addQueryChart" class="table-button">新增設定</button>
						</div>
						<template v-for="query_chart in params_component.query_charts">
							<hr />
							<label for="city">城市</label>
							<!-- TODO: 選擇city時從 option_city 移除被選擇的 city-->
							<select
								v-model="query_chart.city"
								@change="console.log('change evet')"
								name="city"
								id="city"
							>
								<template v-for="city in options_city" :key="city.value">
									<option :value="city.value">{{ city.label }}</option>
								</template>
							</select>
							<label>資料來源*</label>
							<input v-model="query_chart.source" type="text" :minlength="1" :maxlength="12" required />
							<label>資料連結</label>
							<InputTags
								:tags="query_chart.links"
								@deletetag="
									(index) => {
										query_chart.links.splice(index, 1);
									}
								"
								@updatetagorder="
									(updatedTags) => {
										query_chart.links = updatedTags;
									}
								"
							/>
							<input
								v-model="tempInputStorage.link"
								type="text"
								:minlength="1"
								@keypress.enter="
									() => {
										query_chart.links.push(tempInputStorage.link);
										tempInputStorage.link = '';
									}
								"
							/>
							<label>貢獻者</label>
							<InputTags
								:tags="query_chart.contributors"
								@deletetag="
									(index) => {
										query_chart.contributors.splice(index, 1);
									}
								"
								@updatetagorder="
									(updatedTags) => {
										query_chart.contributors = updatedTags;
									}
								"
							/>
							<input
								v-model="tempInputStorage.contributor"
								type="text"
								@keypress.enter="
									() => {
										if (tempInputStorage.contributor.length > 0) {
											query_chart.contributors.push(tempInputStorage.contributor);
											tempInputStorage.contributor = '';
										}
									}
								"
							/>
							<label for="query_chart">圖表SQL* (query_chart)</label>
							<textarea
								v-model="query_chart.query_chart"
								name="query_chart"
								id="query_chart"
								required
							></textarea>
						</template>
					</div>

					<div v-if="currentTab === 1">
						<div class="flex justify-between mt-2">
							<label>組件 Index - {{ params_component.index }}</label>
							<button class="table-button" @click.prevent="addMapConfig">新增地圖</button>
						</div>
						<template v-for="(map, index) in params_maps">
							<div class="admincomponenttemplate-settings-items">
								<hr v-if="Number.parseInt(index) > 0" />
								<label>
									{{
										`地圖${Number.parseInt(index) + 1} 名稱* ${map.title} (${map.title.length}/10)`
									}}
								</label>
								<input v-model="map.title" type="text" :minlength="1" :maxlength="10" required />
								<label>地圖{{ `${Number.parseInt(index) + 1}尺寸* (${map.type.length}/20)` }}</label>
								<input v-model="map.size" type="text" :minlength="1" :maxlength="10" />
								<label>{{ `${Number.parseInt(index) + 1}來源 (${map.source.length}/20)` }}</label>
								<input v-model="map.source" type="text" :minlength="1" :maxlength="20" />
								<label>地圖{{ Number.parseInt(index) + 1 }} 預設變形（大小/圖示）</label>
								<div class="two-block">
									<select v-model="map.type">
										<option :value="''">無</option>
										<option value="small">small (點圖)</option>
										<option value="big">big (點圖)</option>
										<option value="wide">wide (線圖)</option>
									</select>
									<select v-model="map.icon">
										<option :value="''">無</option>
										<option value="heatmap">heatmap (點圖)</option>
										<option value="dash">dash (線圖)</option>
										<option value="metro">metro (符號圖)</option>
										<option value="metro-density">metro-density (符號圖)</option>
										<option value="triangle_green">triangle_green (符號圖)</option>
										<option value="triangle_white">triangle_white (符號圖)</option>
										<option value="youbike">youbike (符號圖)</option>
										<option value="bus">bus (符號圖)</option>
									</select>
								</div>
								<label>地圖{{ Number.parseInt(index) + 1 }} Paint屬性</label>
								<textarea v-model="map.paint" />
								<label>地圖{{ Number.parseInt(index) + 1 }} Popup標籤</label>
								<textarea v-model="map.property" />
								<button
									@click.prevent="params_maps.splice(Number.parseInt(index), 1)"
									class="table-button mt-2"
								>
									刪除
								</button>
							</div>
						</template>
					</div>
					<div v-if="currentTab === 2">
						<div class="admincomponenttemplate-settings-items">
							<div class="admincomponenttemplate-settings-table">
								<form ref @submit.prevent="handleTableSubmit" enctype="multipart/form-data">
									<div class="form_header">
										<div class="form_input">
											<label for="table_name">資料表名稱</label>
											<input
												v-model="params_table.table.name"
												type="text"
												name="table_name"
												id="table_name"
											/>
										</div>
										<div class="table-buttons">
											<button class="table-button" @click.prevent="addColumn">新增欄位</button>
											<label for="file" class="table-button">上傳資料</label>
											<input type="file" name="file" id="file" @change="handleFileChange" />
										</div>
									</div>
									<div>
										<table>
											<thead>
												<tr>
													<th></th>
													<th>名稱*</th>
													<th>類型*</th>
													<th>備註</th>
													<th>是否為主鍵</th>
													<th>是否允許空值</th>
												</tr>
											</thead>
											<tbody>
												<template
													v-for="(column, index) in params_table.table.columns"
													:key="index"
												>
													<tr>
														<td>
															<button
																class="delete-column"
																@click.prevent="
																	() => {
																		params_table.table.columns.splice(index, 1);
																	}
																"
															>
																delete
															</button>
														</td>
														<td>
															<input v-model="column.name" type="text" />
														</td>
														<td>
															<select v-model="column.type">
																<option
																	v-for="type in options_column_types"
																	:key="type.value"
																	:value="type.value"
																>
																	{{ type.label }}
																</option>
															</select>
														</td>

														<td>
															<textarea
																v-model="column.comment"
																name="comment"
																:id="`comment-${index}`"
															></textarea>
														</td>
														<td>
															<input
																v-model="column.is_primary_key"
																type="checkbox"
																:checked="column.is_primary_key"
															/>
														</td>
														<td>
															<input
																v-model="column.notNull"
																type="checkbox"
																:checked="column.notNull"
															/>
														</td>
													</tr>
												</template>
											</tbody>
										</table>
									</div>
								</form>
								<button class="table-button submit-table-button" @click="handleTableSubmit">
									新增資料表
								</button>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</DialogContainer>
</template>

<style scoped lang="scss">
	.admincomponenttemplate {
		width: 750px;
		height: 500px;

		@media (max-width: 770px) {
			display: none;
		}
		@media (max-height: 520px) {
			display: none;
		}

		&-header {
			display: flex;
			justify-content: space-between;

			button {
				display: flex;
				align-items: center;
				justify-self: baseline;
				padding: 2px 4px;
				border-radius: 5px;
				background-color: var(--color-highlight);
				font-size: var(--font-ms);
			}
		}

		&-content {
			border: solid 1px var(--color-border);
			border-radius: 0px 5px 5px 5px;
			height: calc(100% - 70px);
			overflow-y: scroll;
		}

		&-tabs {
			height: 30px;
			display: flex;
			align-items: center;
			margin-top: var(--font-s);

			button {
				// width: 70px;
				height: 30px;
				padding: 3px 10px;
				border-radius: 5px 5px 0px 0px;
				background-color: var(--color-border);
				font-size: var(--font-m);
				color: var(--color-text);
				cursor: pointer;
				transition: background-color 0.2s;

				&:hover {
					background-color: var(--color-complement-text);
				}
			}
			.active {
				background-color: var(--color-complement-text);
			}
		}

		&-settings {
			padding: 0 0.5rem 0.5rem 0.5rem;
			// margin-right: var(--font-ms);

			label {
				margin: 8px 0 4px;
				font-size: var(--font-s);
				color: var(--color-complement-text);
			}

			.two-block {
				display: grid;
				grid-template-columns: 1fr 1fr;
				column-gap: 0.5rem;
			}
			.three-block {
				display: grid;
				grid-template-columns: 1fr 2rem 1fr;
				column-gap: 0.5rem;
			}

			&-items {
				display: flex;
				flex-direction: column;

				hr {
					margin: var(--font-ms) 0 0.5rem;
					border: none;
					border-bottom: dashed 1px var(--color-complement-text);
				}
			}

			&-inputcolor {
				width: 140px;
				height: 40px;
				appearance: none;
				display: flex;
				justify-content: center;
				align-items: center;
				padding: 0;
				outline: none;
				cursor: pointer;

				&::-webkit-color-swatch {
					border: none;
					border-radius: 5px;
				}
				&::-moz-color-swatch {
					border: none;
				}
				&:before {
					content: "選擇顏色";
					position: absolute;
					display: block;
					border-radius: 5px;
					font-size: var(--font-ms);
					color: var(--color-complement-text);
				}
				&:focus:before {
					content: "點擊空白處確認";
					text-shadow: 0px 0px 1px #666;
				}
			}

			.table-button {
				display: flex;
				align-items: center;
				width: 100px;
				height: 40px;
				background-color: var(--color-border);
				text-align: center;
				border: none;
				border-radius: 5px;
				color: var(--color-text);
				cursor: pointer;
				font-size: var(--font-ms);
				transition: background-color 0.2s;
				align-items: center;
				justify-content: center;

				&:hover {
					background-color: var(--color-complement-text);
				}
			}

			&-table {
				padding: 10px 10px;

				.form_header {
					display: flex;
					align-items: end;
				}

				.table-buttons {
					display: flex;
					gap: 10px;
					margin-left: auto;
				}

				.submit-table-button {
					margin-left: auto;
				}
				label.table-button {
					margin: 0;
				}

				.delete-column {
					font-size: 28px;
					font-family: var(--font-icon);
				}

				table {
					border-collapse: collapse;
					width: 100%;
					margin-top: 5px;
				}
				th,
				td {
					font-size: 14px;
					border: 1px solid #555;
					padding: 4px 8px;
					text-align: center; /* 水平置中 */
					vertical-align: middle;
				}

				input[type="checkbox"] {
					display: block;
					margin: auto;
					width: 18px;
					height: 18px;
					// appearance: none;
					border: 1px solid var(--color-border);
					border-radius: 5px;
					cursor: pointer;
					transition:
						background-color 0.2s,
						border-color 0.2s;

					&:checked {
						background-color: var(--color-complement-text);
						border-color: var(--color-complement-text);
					}
				}

				input[type="file"] {
					display: none;
				}
			}

			&::-webkit-scrollbar {
				width: 4px;
			}
			&::-webkit-scrollbar-thumb {
				background-color: rgba(136, 135, 135, 0.5);
				border-radius: 4px;
			}
			&::-webkit-scrollbar-thumb:hover {
				background-color: rgba(136, 135, 135, 1);
			}
		}

		&-preview {
			display: flex;
			flex-direction: column;
			justify-content: center;
			align-items: center;
			border-radius: 5px;
			border: solid 1px var(--color-border);
		}
	}
	.time-block {
		display: flex;
		align-items: center;
		column-gap: 4px;
		select {
			width: 100px;
		}
	}

	.form_input {
		label {
			margin-right: 10px;
		}
	}

	.flex {
		display: flex;
	}

	.justify-between {
		justify-content: space-between;
	}

	.mt-2 {
		margin-top: 8px;
	}
</style>
