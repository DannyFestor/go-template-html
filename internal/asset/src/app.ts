import "./app.css";
import "htmx.org";
import Alpine from "@alpinejs/csp";
import persist from "@alpinejs/persist";

import { disclosure } from "./components/disclosure";
import { nav } from "./components/nav";

Alpine.plugin(persist);
Alpine.data("disclosure", disclosure);
Alpine.data("nav", nav);
Alpine.start();
