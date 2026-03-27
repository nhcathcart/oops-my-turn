import { useQuery } from '@tanstack/react-query'
import { getHelloOptions } from '@/api/generated/@tanstack/react-query.gen'

export function useGetHello() {
  return useQuery(getHelloOptions())
}
