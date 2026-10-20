import "./app.css";
import "htmx.org";
import Alpine from "@alpinejs/csp";
import persist from "@alpinejs/persist";

import { appearance } from "./components/appearance";
import { disclosure } from "./components/disclosure";
import { nav } from "./components/nav";

Alpine.plugin(persist);
Alpine.data("appearance", appearance);
Alpine.data("disclosure", disclosure);
Alpine.data("nav", nav);
Alpine.start();
