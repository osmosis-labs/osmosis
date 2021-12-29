// register Osmosis components
import Reference from "./osmosis-components/Reference";
import CodeSignature from "./osmosis-components/CodeSignature";
import JsonTreeView from "./osmosis-components/JsonTreeView";
import Param from "./osmosis-components/Param";
import TypeDesc from "./osmosis-components/TypeDesc";
import BoxField from "./osmosis-components/BoxField";
import ParamsList from "./osmosis-components/ParamsList";
import DarkModeSwitch from "./osmosis-components/DarkModeSwitch";

export default ({ Vue }) => {
  Vue.component("Reference", Reference);
  Vue.component("CodeSignature", CodeSignature);
  Vue.component("JsonTreeView", JsonTreeView);
  Vue.component("Parameter", Param);
  Vue.component("ParamsList", ParamsList);
  Vue.component("TypeDesc", TypeDesc);
  Vue.component("BoxField", BoxField);
  Vue.component("DarkModeSwitch", DarkModeSwitch);

};
