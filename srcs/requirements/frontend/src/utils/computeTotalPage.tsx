import {tListResponse} from "@/types/api";

export default function computeTotalPage(data?: tListResponse<unknown[]>) {
    if (data?.meta && data.meta.per_page !== 0)
        return Math.ceil(data.meta.total / data.meta.per_page);
    return 1;
}
