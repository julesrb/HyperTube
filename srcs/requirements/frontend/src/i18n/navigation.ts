import {createNavigation} from "next-intl/navigation";
import {routing} from "@/i18n/request";

export const {Link, usePathname, useRouter} = createNavigation(routing);
