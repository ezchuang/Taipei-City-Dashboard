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

const default_params = {
	index: "",
	name: "",
	history_config: null,
	map_config: [null],
	map_filter: null,
	time_from: null,
	time_to: null,
	update_freq: null,
	update_freq_unit: "minute",
	source: "",
	short_desc: "",
	long_desc: "",
	use_case: "",
	links: [],
	contributors: [],
	query_type: "map_legend",
	// columns structure {name: "", type: "", notNull: false, comment: ""}
	schema: {
		table: "",
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
};
const currentTab = ref(0);
const params = ref({ ...default_params });

const column_types = [
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

const tempInputStorage = ref({
	link: "",
	contributor: "",
	chartColor: "#000000",
	historyColor: "#000000",
});

const buttonTabs = ref([
	{ label: "整體", order: 0 },
	{ label: "地圖圖層設定", order: 1 },
	{ label: "資料表設定", order: 2 },
]);

function addColumn() {
	params.value.schema.columns.push({
		is_primary_key: false,
		name: "",
		type: "text",
		description: "",
		notNull: false,
		comment: "",
	});
}

function handleClose() {
	dialogStore.hideAllDialogs();
	params.value = { ...default_params };
}

async function handleSubmit() {
	const {
		index,
		name,
		source,
		update_freq,
		update_freq_unit,
		short_desc,
		long_desc,
		use_case,
	} = params.value;
	if (
		!index ||
		!name ||
		!source ||
		(!update_freq && update_freq !== 0) ||
		!update_freq_unit ||
		!short_desc ||
		!long_desc ||
		!use_case
	) {
		dialogStore.showNotification(
			"fail",
			"新建失敗，請確實填寫所有必填欄位",
		);
		return;
	}
	try {
		const createRes = await http.post(`/component/`, params.value);
		const getRes = await http.get(`/component/${createRes.data.data.id}`);
		adminStore.getComponentData(getRes.data.data);
		dialogStore.showNotification("success", "新建組建成功");
		dialogStore.hideAllDialogs();
		dialogStore.showDialog("adminComponentSettings");
	} catch (error) {
		console.error(error);
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
	<DialogContainer
		:dialog="`adminAddComponentTemplate`"
		@on-close="handleClose"
	>
		<div class="admincomponenttemplate">
			<div class="admincomponenttemplate-header">
				<h2>組件設定</h2>
				<button @click="handleSubmit">確定新增</button>
			</div>
			<div class="admincomponenttemplate-tabs">
				<template v-for="tab in buttonTabs" :key="tab.order">
					<button
						:class="currentTab === tab.order ? 'active' : ''"
						@click="currentTab = tab.order"
					>
						{{ tab.label }}
					</button>
				</template>
			</div>
			<div class="admincomponenttemplate-content">
				<!-- TODO: 下面的 input 由於寬度變寬，有些 input 要放到同一排 input 才不會太長 -->
				<div class="admincomponenttemplate-settings">
					<div
						v-if="currentTab === 0"
						class="admincomponenttemplate-settings-items"
					>
						<label>組件名稱* ({{ params.name.length }}/10)</label>
						<input
							v-model="params.name"
							type="text"
							:minlength="1"
							:maxlength="10"
							required
						/>
						<div class="two-block">
							<label>組件 Index*</label>
						</div>
						<div class="two-block">
							<input
								v-model="params.index"
								type="text"
								required
							/>
						</div>
						<label>資料來源*</label>
						<input
							v-model="params.source"
							type="text"
							:minlength="1"
							:maxlength="12"
							required
						/>
						<label>更新頻率* (0 = 不定期更新){{}}</label>
						<div class="two-block">
							<input
								v-model="params.update_freq"
								type="number"
								:min="0"
								:max="31"
								required
							/>
							<select v-model="params.update_freq_unit">
								<option value="minute">分</option>
								<option value="hour">時</option>
								<option value="day">天</option>
								<option value="week">週</option>
								<option value="month">月</option>
								<option value="year">年</option>
							</select>
						</div>
						<label>資料區間</label>
						<div class="time-block">
							<select v-model="params.time_from" required>
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
							<span v-show="isShowTimeToBlock(params.time_from)"
								>～</span
							>
							<select
								v-show="isShowTimeToBlock(params.time_from)"
								v-model="params.time_to"
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
						<label required>組件簡述* ({{}}/50)</label>
						<textarea
							v-model="params.short_desc"
							:minlength="1"
							:maxlength="50"
							required
						/>
						<label>組件詳述* ({{}}/100)</label>
						<textarea
							v-model="params.long_desc"
							:minlength="1"
							:maxlength="100"
							required
						/>
						<label>範例情境* ({{}}/100)</label>
						<textarea
							v-model="params.use_case"
							:minlength="1"
							:maxlength="100"
							required
						/>
						<label>資料連結</label>
						<InputTags
							:tags="params.links"
							@deletetag="
								(index) => {
									params.links.splice(index, 1);
								}
							"
							@updatetagorder="
								(updatedTags) => {
									params.links = updatedTags;
								}
							"
						/>
						<input
							v-model="tempInputStorage.link"
							type="text"
							:minlength="1"
							@keypress.enter="
								() => {
									if (tempInputStorage.link.length > 0) {
										params.links.push(
											tempInputStorage.link,
										);
										tempInputStorage.link = '';
									}
								}
							"
						/>
						<label>貢獻者</label>
						<InputTags
							:tags="params.contributors"
							@deletetag="
								(index) => {
									params.contributors.splice(index, 1);
								}
							"
							@updatetagorder="
								(updatedTags) => {
									params.contributors = updatedTags;
								}
							"
						/>
						<input
							v-model="tempInputStorage.contributor"
							type="text"
							@keypress.enter="
								() => {
									if (
										tempInputStorage.contributor.length > 0
									) {
										params.contributors.push(
											tempInputStorage.contributor,
										);
										tempInputStorage.contributor = '';
									}
								}
							"
						/>
					</div>

					<div v-if="currentTab === 1">
						<div
							v-for="(_map_config, index) in 2"
							:key="index"
							class="admincomponenttemplate-settings-items"
						>
							<hr v-if="index > 0" />
							<label>地圖{{ index + 1 }} ID / Index</label>

							<label>地圖{{ index + 1 }} 名稱* ({{}}/10)</label>
							<input
								type="text"
								:minlength="1"
								:maxlength="10"
								required
							/>
							<label>地圖{{ index + 1 }} 類型*</label>
							<label
								>地圖{{
									index + 1
								}}
								預設變形（大小/圖示）</label
							>
							<div class="two-block">
								<select>
									<option :value="''">無</option>
									<option value="small">small (點圖)</option>
									<option value="big">big (點圖)</option>
									<option value="wide">wide (線圖)</option>
								</select>
								<select>
									<option :value="''">無</option>
									<option value="heatmap">
										heatmap (點圖)
									</option>
									<option value="dash">dash (線圖)</option>
									<option value="metro">
										metro (符號圖)
									</option>
									<option value="metro-density">
										metro-density (符號圖)
									</option>
									<option value="triangle_green">
										triangle_green (符號圖)
									</option>
									<option value="triangle_white">
										triangle_white (符號圖)
									</option>
									<option value="youbike">
										youbike (符號圖)
									</option>
									<option value="bus">bus (符號圖)</option>
								</select>
							</div>
							<label>地圖{{ index + 1 }} Paint屬性</label>
							<textarea />
							<label>地圖{{ index + 1 }} Popup標籤</label>
							<textarea />
						</div>
					</div>
					<div v-if="currentTab === 2">
						<div class="admincomponenttemplate-settings-items">
							<div class="admincomponenttemplate-settings-schema">
								<div class="two-block">
									<label>建立資料表</label>
									<button
										class="add-column"
										@click="addColumn"
									>
										新增欄位
									</button>
								</div>
								<div>
									<table>
										<thead>
											<tr>
												<th></th>
												<th>名稱</th>
												<th>類型</th>
												<th>備註</th>
												<th>是否為主鍵</th>
												<th>是否允許空值</th>
											</tr>
										</thead>
										<tbody>
											<template
												v-for="(column, index) in params
													.schema.columns"
												:key="index"
											>
												<tr>
													<td>
														<button
															class="delete-column"
															@click="
																() => {
																	params.schema.columns.splice(
																		index,
																		1,
																	);
																}
															"
														>
															delete
														</button>
													</td>
													<td>
														<input type="text" />
													</td>
													<td>
														<select
															v-model="
																column.type
															"
														>
															<option
																v-for="type in column_types"
																:key="
																	type.value
																"
																:value="
																	type.value
																"
															>
																{{ type.label }}
															</option>
														</select>
													</td>

													<td>
														<textarea
															v-model="
																column.comment
															"
															name="comment"
															:id="`comment-${index}`"
														></textarea>
													</td>
													<td>
														<input
															v-model="
																column.is_primary_key
															"
															type="checkbox"
															:checked="
																column.is_primary_key
															"
														/>
													</td>
													<td>
														<input
															v-model="
																column.notNull
															"
															type="checkbox"
															:checked="
																column.notNull
															"
														/>
													</td>
												</tr>
											</template>
										</tbody>
									</table>
								</div>
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

		&-schema {
			padding: 10px 10px;
			table {
				border-collapse: collapse; /* 邊框合併，避免雙邊框 */
				width: 100%; /* 可選：讓表格撐滿 */
			}

			.add-column {
				width: 100px;
				margin-left: auto;
				background-color: var(--color-border);
				border: none;
				border-radius: 5px;
				color: var(--color-text);
				cursor: pointer;
				padding: 2px 4px;
				font-size: var(--font-ms);
				transition: background-color 0.2s;

				&:hover {
					background-color: var(--color-complement-text);
				}
			}

			.delete-column {
				font-size: 28px;
				font-family: var(--font-icon);
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

			table {
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
</style>
