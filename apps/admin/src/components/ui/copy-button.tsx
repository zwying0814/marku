import { Check, Copy } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import { Button, type ButtonProps } from '@/components/ui/button'

type CopyButtonProps = Omit<ButtonProps, 'onClick'> & {
  value: string
  label?: string
  showLabel?: boolean
}

export function CopyButton({
  value,
  label = '复制',
  showLabel = true,
  size = 'sm',
  variant = 'outline',
  ...props
}: CopyButtonProps) {
  const [copied, setCopied] = useState(false)

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      toast.success('已复制到剪贴板')
      window.setTimeout(() => setCopied(false), 1600)
    } catch {
      toast.error('复制失败，请手动选择文本复制')
    }
  }

  return (
    <Button
      type="button"
      size={size}
      variant={variant}
      onClick={handleCopy}
      aria-label={label}
      {...props}
    >
      {copied ? <Check className="size-3.5" /> : <Copy className="size-3.5" />}
      {showLabel ? (copied ? '已复制' : label) : null}
    </Button>
  )
}
