import {QueryClient, useQuery} from "@tanstack/react-query";
import {useLocale} from "next-intl";
import {iApplication, tListResponse} from "@/types/api";
import {iComment} from "@/types/comment";

export default function useApiQuery<T>(key: unknown[], fn: (locale: string, signal?: AbortSignal) => Promise<T>, enabled = true) {
    const locale = useLocale();
    const queryKey = [...key, locale];

    return useQuery({
        queryKey: queryKey,
        queryFn: ({signal}) => fn(locale, signal),
        enabled: enabled && !!locale,
        retry: false
    });
}

export function updateTotal(meta: tListResponse<unknown[]>["meta"], delta: number) {
    if (!meta)
        return {page: 0, per_page: 12, total: 1};
    return {...meta, total: Math.max(0, meta.total + delta)};
}

export function addQuery<T>(queryClient: QueryClient, key: unknown[], newContent: T) {
    const queries = queryClient.getQueriesData<tListResponse<unknown[]>>({queryKey: key});
    queries.forEach(([queryKey, current]) => {
        if (!current)
            return;
        queryClient.setQueryData(queryKey, {
            ...current,
            data: [newContent, ...current.data],
            meta: updateTotal(current.meta, 1),
        });
    });
}

export function updateQuery<T extends iApplication | iComment>(queryClient: QueryClient, key: unknown[], newContent: T) {
    const queries = queryClient.getQueriesData<tListResponse<T[]>>({queryKey: key});
    queries.forEach(([queryKey, current]) => {
        if (!current)
            return;
        queryClient.setQueryData(queryKey, {
            ...current,
            data: current.data.map((v) => {
                if (v.id === newContent.id)
                    return newContent;
                return v;
            })
        });
    });
}

export function removeQuery(queryClient: QueryClient, key: unknown[], deleteObjId: number) {
    const queries = queryClient.getQueriesData<tListResponse<unknown[]>>({queryKey: key});
    queries.forEach(([queryKey, current]) => {
        if (!current)
            return;
        const nextData = current.data.filter((i) => {
            const data = i as iApplication | iComment;
            return data.id !== deleteObjId
        })
        if (nextData.length === current.data.length)
            return;
        queryClient.setQueryData(queryKey, {
            ...current,
            data: nextData,
            meta: updateTotal(current.meta, -1),
        });
    });
}
